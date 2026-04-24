package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ── server ────────────────────────────────────────────────────────────────────

// server holds the absolute path of the media root and the HTTP multiplexer.
type server struct {
	root string        // absolute path to the directory being served
	mux  *http.ServeMux
}

// newServer constructs a server rooted at root and registers all API and
// static routes on a fresh ServeMux.
func newServer(root string) *server {
	s := &server{root: root, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /favicon.svg", s.handleFavicon)
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/videos", s.handleVideos)
	s.mux.HandleFunc("/api/folders", s.handleFolders)
	s.mux.HandleFunc("/api/mkdir", s.handleMkdir)
	s.mux.HandleFunc("/api/rmdir", s.handleRmdir)
	s.mux.HandleFunc("/api/move", s.handleMove)
	s.mux.HandleFunc("/api/upload", s.handleUpload)
	s.mux.HandleFunc("/video", s.handleVideo)
	return s
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

func (s *server) handleVideos(w http.ResponseWriter, r *http.Request) {
	videos, err := scanVideos(s.root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

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

// folderInfo describes a subdirectory under the media root.
// It is serialised to JSON and returned by the /api/folders endpoint.
type folderInfo struct {
	Name          string `json:"name"`           // slash-separated path relative to root
	HasOtherFiles bool   `json:"has_other_files"` // true if the directory contains non-video files
}

func (s *server) handleFolders(w http.ResponseWriter, r *http.Request) {
	seen := map[string]*folderInfo{}
	filepath.WalkDir(s.root, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == s.root {
			return err
		}
		rel, _ := filepath.Rel(s.root, path)
		slashRel := filepath.ToSlash(rel)
		if d.IsDir() {
			if _, exists := seen[slashRel]; !exists {
				seen[slashRel] = &folderInfo{Name: slashRel}
			}
		} else {
			ext := strings.ToLower(filepath.Ext(path))
			if !videoExts[ext] {
				parent := filepath.ToSlash(filepath.Dir(rel))
				if parent != "." {
					if fi, exists := seen[parent]; exists {
						fi.HasOtherFiles = true
					} else {
						seen[parent] = &folderInfo{Name: parent, HasOtherFiles: true}
					}
				}
			}
		}
		return nil
	})
	folders := make([]folderInfo, 0, len(seen))
	for _, fi := range seen {
		folders = append(folders, *fi)
	}
	sort.Slice(folders, func(i, j int) bool { return folders[i].Name < folders[j].Name })
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folders)
}

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
	clean := filepath.Clean(filepath.FromSlash(req.Folder))
	if strings.HasPrefix(clean, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rootAbs, _ := filepath.Abs(s.root)
	destAbs, err := filepath.Abs(filepath.Join(s.root, clean))
	if err != nil || !strings.HasPrefix(destAbs, rootAbs+string(os.PathSeparator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := os.MkdirAll(destAbs, 0755); err != nil {
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
	clean := filepath.Clean(filepath.FromSlash(req.Folder))
	if clean == "." || strings.HasPrefix(clean, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rootAbs, _ := filepath.Abs(s.root)
	dirAbs, err := filepath.Abs(filepath.Join(s.root, clean))
	if err != nil || !strings.HasPrefix(dirAbs, rootAbs+string(os.PathSeparator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Move all files inside the directory to root
	err = filepath.WalkDir(dirAbs, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		dest := filepath.Join(s.root, filepath.Base(path))
		// Avoid overwriting existing files at root
		if _, statErr := os.Stat(dest); statErr == nil {
			ext := filepath.Ext(filepath.Base(path))
			base := strings.TrimSuffix(filepath.Base(path), ext)
			dest = filepath.Join(s.root, base+"_1"+ext)
		}
		return os.Rename(path, dest)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.RemoveAll(dirAbs); err != nil {
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

	srcClean := filepath.Clean(filepath.FromSlash(req.Path))
	if strings.HasPrefix(srcClean, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rootAbs, _ := filepath.Abs(s.root)
	srcAbs, err := filepath.Abs(filepath.Join(s.root, srcClean))
	if err != nil || !strings.HasPrefix(srcAbs, rootAbs+string(os.PathSeparator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	destFolder := filepath.Clean(filepath.FromSlash(req.DestFolder))
	if destFolder == "." || destFolder == "/" {
		destFolder = ""
	}
	if strings.HasPrefix(destFolder, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	destDir := filepath.Join(s.root, destFolder)
	destDirAbs, err := filepath.Abs(destDir)
	if err != nil || (!strings.HasPrefix(destDirAbs, rootAbs+string(os.PathSeparator)) && destDirAbs != rootAbs) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		http.Error(w, "cannot create destination", http.StatusInternalServerError)
		return
	}
	destAbs := filepath.Join(destDir, filepath.Base(srcAbs))
	if err := os.Rename(srcAbs, destAbs); err != nil {
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
