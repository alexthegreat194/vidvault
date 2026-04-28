package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── server ────────────────────────────────────────────────────────────────────

// server holds the absolute path of the media root and the HTTP multiplexer.
type server struct {
	root        string // absolute path to the directory being served
	favorites   *FavoritesStore
	tags        *TagsStore
	collections *WatchCollectionsStore
	videoCache  *videoScanCache
	mux         *http.ServeMux
}

var serverLog = fileLogger("server")

// newServer constructs a server rooted at root and registers all API and
// static routes on a fresh ServeMux.
func newServer(root string) (*server, error) {
	logDebug(serverLog, "initializing server", "root", root)
	favorites, err := newFavoritesStore()
	if err != nil {
		return nil, err
	}
	tags, err := newTagsStore()
	if err != nil {
		return nil, err
	}
	collections, err := newWatchCollectionsStore()
	if err != nil {
		return nil, err
	}

	s := &server{
		root:        root,
		favorites:   favorites,
		tags:        tags,
		collections: collections,
		videoCache:  newVideoScanCache(),
		mux:         http.NewServeMux(),
	}
	s.mux.HandleFunc("GET /favicon.svg", s.handleFavicon)
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/videos", s.handleVideos)
	s.mux.HandleFunc("GET /api/videos/stream", s.handleVideosStream)
	s.mux.HandleFunc("GET /api/favorites", s.handleFavorites)
	s.mux.HandleFunc("POST /api/favorites/set", s.handleSetFavorite)
	s.mux.HandleFunc("GET /api/tags", s.handleTags)
	s.mux.HandleFunc("POST /api/tags/create", s.handleCreateTag)
	s.mux.HandleFunc("POST /api/tags/assign", s.handleAssignTag)
	s.mux.HandleFunc("POST /api/tags/delete", s.handleDeleteTag)
	s.mux.HandleFunc("GET /api/collections", s.handleCollections)
	s.mux.HandleFunc("POST /api/collections/create", s.handleCreateCollection)
	s.mux.HandleFunc("POST /api/collections/rename", s.handleRenameCollection)
	s.mux.HandleFunc("POST /api/collections/delete", s.handleDeleteCollection)
	s.mux.HandleFunc("POST /api/collections/videos/set", s.handleSetCollectionVideo)
	s.mux.HandleFunc("POST /api/collections/videos/bulk", s.handleBulkCollectionVideos)
	s.mux.HandleFunc("/api/folders", s.handleFolders)
	s.mux.HandleFunc("/api/mkdir", s.handleMkdir)
	s.mux.HandleFunc("/api/rmdir", s.handleRmdir)
	s.mux.HandleFunc("/api/move", s.handleMove)
	s.mux.HandleFunc("/api/delete", s.handleDelete)
	s.mux.HandleFunc("/api/upload", s.handleUpload)
	s.mux.HandleFunc("/video", s.handleVideo)
	serverLog.Info("routes registered", "root", root)
	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
	logDebug(serverLog, "request started", "method", r.Method, "path", r.URL.Path, "query", r.URL.RawQuery, "remote", r.RemoteAddr)
	s.mux.ServeHTTP(rw, r)
	fields := []any{
		"method", r.Method,
		"path", r.URL.Path,
		"status", rw.status,
		"duration_ms", time.Since(start).Milliseconds(),
	}
	if debugMode.Load() {
		fields = append(fields, "bytes", rw.bytes, "remote", r.RemoteAddr, "user_agent", r.UserAgent())
	}
	serverLog.Info("request completed", fields...)
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func (s *server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	logDebug(serverLog, "serving favicon")
	data, err := webFS.ReadFile("web/favicon.svg")
	if err != nil {
		serverLog.Error("favicon asset missing", "error", err)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(data)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		logDebug(serverLog, "index route not found", "path", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, indexHTML)
}

// handleVideos HTTP handler for all video metadata
func (s *server) handleVideos(w http.ResponseWriter, r *http.Request) {
	logDebug(serverLog, "handling videos request")
	videos, err := scanVideos(s.root, s.videoCache)
	if err != nil {
		serverLog.Error("failed to scan videos", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i := range videos {
		videos[i].Favorite = s.favorites.IsFavorite(videos[i].Hash)
		videos[i].Tags = s.tags.TagsForHash(videos[i].Hash)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
	logDebug(serverLog, "videos response sent", "count", len(videos))
}

func (s *server) applyVideoMetadata(video *Video) {
	video.Favorite = s.favorites.IsFavorite(video.Hash)
	video.Tags = s.tags.TagsForHash(video.Hash)
}

func (s *server) handleVideosStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	enc := json.NewEncoder(w)
	writeEvent := func(event string, payload any) error {
		frame := map[string]any{
			"type": event,
		}
		if payload != nil {
			frame["payload"] = payload
		}
		if err := enc.Encode(frame); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	start := time.Now()
	if err := writeEvent("scan_start", map[string]any{
		"started_at": start.UTC().Format(time.RFC3339Nano),
	}); err != nil {
		return
	}

	count := 0
	_, err := scanVideosWithCallback(s.root, s.videoCache, func(video Video) error {
		select {
		case <-r.Context().Done():
			return r.Context().Err()
		default:
		}
		s.applyVideoMetadata(&video)
		count++
		return writeEvent("video", video)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		serverLog.Error("failed streaming videos scan", "error", err)
		writeEvent("scan_error", map[string]any{"message": err.Error()})
		return
	}
	_ = writeEvent("scan_complete", map[string]any{
		"count":       count,
		"duration_ms": time.Since(start).Milliseconds(),
	})
}

func (s *server) handleFavorites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.favorites.Snapshot())
}

func (s *server) handleSetFavorite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Hash     string `json:"hash"`
		Favorite bool   `json:"favorite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing set favorite payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Hash) == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.favorites.Set(req.Hash, req.Favorite); err != nil {
		serverLog.Error("failed updating favorite", "hash", req.Hash, "favorite", req.Favorite, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logDebug(serverLog, "favorite updated", "hash", req.Hash, "favorite", req.Favorite)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.tags.Snapshot())
}

func (s *server) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing create tag payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	tag, err := s.tags.Create(req.Name)
	if err != nil {
		serverLog.Error("failed creating tag", "name", req.Name, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
	logDebug(serverLog, "tag created", "tag_id", tag.ID, "name", tag.Name)
}

func (s *server) handleAssignTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Hash     string `json:"hash"`
		TagID    string `json:"tag_id"`
		Assigned bool   `json:"assigned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing assign tag payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.tags.SetAssignment(req.Hash, req.TagID, req.Assigned); err != nil {
		serverLog.Error("failed assigning tag", "hash", req.Hash, "tag_id", req.TagID, "assigned", req.Assigned, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logDebug(serverLog, "tag assignment updated", "hash", req.Hash, "tag_id", req.TagID, "assigned", req.Assigned)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagID string `json:"tag_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing delete tag payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.tags.Delete(req.TagID); err != nil {
		serverLog.Error("failed deleting tag", "tag_id", req.TagID, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logDebug(serverLog, "tag deleted", "tag_id", req.TagID)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleCollections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Collections []WatchCollection `json:"collections"`
	}{
		Collections: s.collections.Snapshot(),
	})
}

func (s *server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing create collection payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	collection, err := s.collections.Create(req.Name)
	if err != nil {
		serverLog.Error("failed creating collection", "name", req.Name, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(collection)
	logDebug(serverLog, "collection created", "id", collection.ID, "name", collection.Name)
}

func (s *server) handleRenameCollection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing rename collection payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.collections.Rename(req.ID, req.Name); err != nil {
		serverLog.Error("failed renaming collection", "id", req.ID, "name", req.Name, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing delete collection payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.collections.Delete(req.ID); err != nil {
		serverLog.Error("failed deleting collection", "id", req.ID, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleSetCollectionVideo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string `json:"id"`
		Hash     string `json:"hash"`
		Assigned bool   `json:"assigned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing set collection video payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.collections.SetVideo(req.ID, req.Hash, req.Assigned); err != nil {
		serverLog.Error("failed setting collection video", "id", req.ID, "hash", req.Hash, "assigned", req.Assigned, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleBulkCollectionVideos(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string   `json:"id"`
		Hashes []string `json:"hashes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("failed parsing bulk collection payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.collections.AddVideos(req.ID, req.Hashes); err != nil {
		serverLog.Error("failed bulk adding collection videos", "id", req.ID, "hash_count", len(req.Hashes), "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Handler for
func (s *server) handleVideo(w http.ResponseWriter, r *http.Request) {
	logDebug(serverLog, "video request received", "query", r.URL.RawQuery)
	relPath, err := url.QueryUnescape(r.URL.Query().Get("path"))
	if err != nil || relPath == "" {
		serverLog.Error("missing or invalid video path", "error", err)
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	// Sanitise: no directory traversal
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if strings.HasPrefix(clean, "..") {
		serverLog.Error("video request forbidden, traversal detected", "path", relPath)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	abs := filepath.Join(s.root, clean)

	// Double-check the resolved path stays inside root
	rootAbs, _ := filepath.Abs(s.root)
	fileAbs, err := filepath.Abs(abs)
	if err != nil || !strings.HasPrefix(fileAbs, rootAbs+string(os.PathSeparator)) {
		serverLog.Error("video request resolved outside root", "path", relPath, "resolved", fileAbs, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	f, err := os.Open(fileAbs)
	if err != nil {
		serverLog.Error("video not found", "path", relPath, "resolved", fileAbs, "error", err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(fileAbs))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Accept-Ranges", "bytes")

	// Serve with range support (needed for video scrubbing)
	fi, _ := f.Stat()
	logDebug(serverLog, "serving video file", "path", relPath, "resolved", fileAbs, "size", fi.Size(), "mime", mimeType)
	http.ServeContent(w, r, filepath.Base(fileAbs), fi.ModTime(), f)
}

// Http Handler for reading all the directories in a root directory
func (s *server) handleFolders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	folders, err := getFolderMetadata(s.root)
	if err != nil {
		serverLog.Error("failed loading folder metadata", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(folders)
	logDebug(serverLog, "folder metadata response sent", "count", len(folders))
}

// Http Handler for making a new directory
func (s *server) handleMkdir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Folder string `json:"folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Folder == "" {
		serverLog.Error("invalid mkdir payload", "folder", req.Folder, "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := makeDirectory(s.root, req.Folder)
	if err != nil {
		serverLog.Error("mkdir failed", "folder", req.Folder, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logDebug(serverLog, "directory created", "folder", req.Folder)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleRmdir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Folder string `json:"folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Folder == "" {
		serverLog.Error("invalid rmdir payload", "folder", req.Folder, "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := removeDirectory(s.root, req.Folder); err != nil {
		serverLog.Error("rmdir failed", "folder", req.Folder, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logDebug(serverLog, "directory removed", "folder", req.Folder)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path       string `json:"path"`
		DestFolder string `json:"dest_folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("invalid move payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := moveFileToDirectory(s.root, req.Path, req.DestFolder); err != nil {
		serverLog.Error("move failed", "path", req.Path, "dest", req.DestFolder, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logDebug(serverLog, "file moved", "path", req.Path, "dest", req.DestFolder)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(512 << 20); err != nil {
		serverLog.Error("failed parsing multipart upload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	folder := filepath.Clean(filepath.FromSlash(r.FormValue("folder")))
	if folder == "." || folder == "/" {
		folder = ""
	}
	rootAbs, _ := filepath.Abs(s.root)
	destDir := filepath.Join(s.root, folder)
	destDirAbs, err := filepath.Abs(destDir)
	if err != nil || (!strings.HasPrefix(destDirAbs, rootAbs+string(os.PathSeparator)) && destDirAbs != rootAbs) {
		serverLog.Error("upload destination forbidden", "folder", folder, "resolved", destDirAbs, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		serverLog.Error("failed creating upload destination", "dest_dir", destDir, "error", err)
		http.Error(w, "cannot create destination", http.StatusInternalServerError)
		return
	}

	type result struct {
		Name  string `json:"name"`
		Error string `json:"error,omitempty"`
	}
	var results []result

	files := r.MultipartForm.File["file"]
	logDebug(serverLog, "processing upload files", "count", len(files), "dest_dir", destDir)
	for _, fh := range files {
		res := uploadOne(fh, destDir)
		results = append(results, result{Name: fh.Filename, Error: res})
		if res == "" {
			logDebug(serverLog, "upload succeeded", "file", fh.Filename)
		} else {
			serverLog.Error("upload failed", "file", fh.Filename, "error", res)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Error("invalid delete payload", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := deleteVideoByPath(s.root, req.Path); err != nil {
		serverLog.Error("delete video failed", "path", req.Path, "error", err)
		switch {
		case errors.Is(err, errEmptyVideoPath):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, errForbiddenPath):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, errVideoNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, errNotAFile):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	logDebug(serverLog, "video deleted via api", "path", req.Path)
	w.WriteHeader(http.StatusOK)
}

// uploadOne copies a single multipart file into destDir, validating its
// extension against videoExts. Returns an error string on failure, or "" on
// success.
func uploadOne(fh *multipart.FileHeader, destDir string) string {
	name := filepath.Base(fh.Filename)
	ext := strings.ToLower(filepath.Ext(name))
	if !videoExts[ext] {
		logDebug(serverLog, "upload rejected due to unsupported extension", "file", fh.Filename, "ext", ext)
		return "unsupported type"
	}
	src, err := fh.Open()
	if err != nil {
		return err.Error()
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(destDir, name))
	if err != nil {
		return err.Error()
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err.Error()
	}
	logDebug(serverLog, "uploaded file saved", "file", name, "dest_dir", destDir)
	return ""
}
