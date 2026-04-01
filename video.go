package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ── supported extensions ─────────────────────────────────────────────────────

// videoExts is the set of file extensions (lowercase, dot-prefixed) recognized
// as video files. Referenced by scanVideos, handleFolders, and uploadOne.
var videoExts = map[string]bool{
	".webm": true,
	".mp4":  true,
	".mpv":  true,
	".mkv":  true,
	".mov":  true,
	".avi":  true,
	".m4v":  true,
	".ogv":  true,
}

// ── video metadata ────────────────────────────────────────────────────────────

// Video represents a single video file discovered under the root directory.
// It is serialised to JSON and returned by the /api/videos endpoint.
type Video struct {
	Name   string `json:"name"`   // base filename, e.g. "clip.mp4"
	Path   string `json:"path"`   // slash-separated path relative to root
	Folder string `json:"folder"` // parent directory relative to root, "/" for root-level files
	Ext    string `json:"ext"`    // lowercase extension including dot, e.g. ".mp4"
}

// scanVideos walks root recursively and returns all video files found, sorted
// by their relative path. Files with extensions not in videoExts are skipped.
func scanVideos(root string) ([]Video, error) {
	var videos []Video
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !videoExts[ext] {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		folder := filepath.Dir(rel)
		if folder == "." {
			folder = "/"
		}
		videos = append(videos, Video{
			Name:   filepath.Base(path),
			Path:   filepath.ToSlash(rel),
			Folder: filepath.ToSlash(folder),
			Ext:    ext,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Path < videos[j].Path
	})
	return videos, nil
}
