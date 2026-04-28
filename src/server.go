package main

import (
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
)

// ── server ────────────────────────────────────────────────────────────────────

// server holds the absolute path of the media root and the HTTP multiplexer.
type server struct {
	root      string // absolute path to the directory being served
	favorites *FavoritesStore
	tags      *TagsStore
	mux       *http.ServeMux
}

// newServer constructs a server rooted at root and registers all API and
// static routes on a fresh ServeMux.
func newServer(root string) (*server, error) {
	favorites, err := newFavoritesStore()
	if err != nil {
		return nil, err
	}
	tags, err := newTagsStore()
	if err != nil {
		return nil, err
	}

	s := &server{root: root, favorites: favorites, tags: tags, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /favicon.svg", s.handleFavicon)
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/videos", s.handleVideos)
	s.mux.HandleFunc("GET /api/favorites", s.handleFavorites)
	s.mux.HandleFunc("POST /api/favorites/set", s.handleSetFavorite)
	s.mux.HandleFunc("GET /api/tags", s.handleTags)
	s.mux.HandleFunc("POST /api/tags/create", s.handleCreateTag)
	s.mux.HandleFunc("POST /api/tags/assign", s.handleAssignTag)
	s.mux.HandleFunc("POST /api/tags/delete", s.handleDeleteTag)
	s.mux.HandleFunc("/api/folders", s.handleFolders)
	s.mux.HandleFunc("/api/mkdir", s.handleMkdir)
	s.mux.HandleFunc("/api/rmdir", s.handleRmdir)
	s.mux.HandleFunc("/api/move", s.handleMove)
	s.mux.HandleFunc("/api/delete", s.handleDelete)
	s.mux.HandleFunc("/api/upload", s.handleUpload)
	s.mux.HandleFunc("/video", s.handleVideo)
	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	data, err := webFS.ReadFile("web/favicon.svg")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(data)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, indexHTML)
}

// handleVideos HTTP handler for all video metadata
func (s *server) handleVideos(w http.ResponseWriter, r *http.Request) {
	videos, err := scanVideos(s.root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i := range videos {
		videos[i].Favorite = s.favorites.IsFavorite(videos[i].Hash)
		videos[i].Tags = s.tags.TagsForHash(videos[i].Hash)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
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
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Hash) == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.favorites.Set(req.Hash, req.Favorite); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	tag, err := s.tags.Create(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

func (s *server) handleAssignTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Hash     string `json:"hash"`
		TagID    string `json:"tag_id"`
		Assigned bool   `json:"assigned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.tags.SetAssignment(req.Hash, req.TagID, req.Assigned); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagID string `json:"tag_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := s.tags.Delete(req.TagID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Handler for
func (s *server) handleVideo(w http.ResponseWriter, r *http.Request) {
	relPath, err := url.QueryUnescape(r.URL.Query().Get("path"))
	if err != nil || relPath == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	// Sanitise: no directory traversal
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if strings.HasPrefix(clean, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	abs := filepath.Join(s.root, clean)

	// Double-check the resolved path stays inside root
	rootAbs, _ := filepath.Abs(s.root)
	fileAbs, err := filepath.Abs(abs)
	if err != nil || !strings.HasPrefix(fileAbs, rootAbs+string(os.PathSeparator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	f, err := os.Open(fileAbs)
	if err != nil {
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
	http.ServeContent(w, r, filepath.Base(fileAbs), fi.ModTime(), f)
}

// Http Handler for reading all the directories in a root directory
func (s *server) handleFolders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	folders, err := getFolderMetadata(s.root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(folders)
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
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := makeDirectory(s.root, req.Folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := removeDirectory(s.root, req.Folder); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := moveFileToDirectory(s.root, req.Path, req.DestFolder); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(512 << 20); err != nil {
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
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		http.Error(w, "cannot create destination", http.StatusInternalServerError)
		return
	}

	type result struct {
		Name  string `json:"name"`
		Error string `json:"error,omitempty"`
	}
	var results []result

	for _, fh := range r.MultipartForm.File["file"] {
		res := uploadOne(fh, destDir)
		results = append(results, result{Name: fh.Filename, Error: res})
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
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := deleteVideoByPath(s.root, req.Path); err != nil {
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
	w.WriteHeader(http.StatusOK)
}

// uploadOne copies a single multipart file into destDir, validating its
// extension against videoExts. Returns an error string on failure, or "" on
// success.
func uploadOne(fh *multipart.FileHeader, destDir string) string {
	name := filepath.Base(fh.Filename)
	ext := strings.ToLower(filepath.Ext(name))
	if !videoExts[ext] {
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
	return ""
}
