package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

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

// Video represents a single video file discovered under the root directory.
// It is serialised to JSON and returned by the /api/videos endpoint.
type Video struct {
	Name     string    `json:"name"`     // base filename, e.g. "clip.mp4"
	Path     string    `json:"path"`     // slash-separated path relative to root
	Folder   string    `json:"folder"`   // parent directory relative to root, "/" for root-level files
	Ext      string    `json:"ext"`      // lowercase extension including dot, e.g. ".mp4"
	Size     int64     `json:"size"`     // file size in bytes
	Modified time.Time `json:"modified"` // file last modified timestamp
	Hash     string    `json:"hash"`     // SHA-256 hash of the file
}

var (
	errEmptyVideoPath = errors.New("missing path")
	errForbiddenPath  = errors.New("forbidden")
	errVideoNotFound  = errors.New("video not found")
	errNotAFile       = errors.New("path is not a file")
)

func isValidVideoExtention(ext string) bool {
	return !videoExts[ext] == false
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Walks root recursively and returns all video files found, sorted
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
		info, err := d.Info()
		if err != nil {
			return err
		}

		hash, err := hashFile(path)
		if err != nil {
			return err
		}
		videos = append(videos, Video{
			Name:     filepath.Base(path),
			Path:     filepath.ToSlash(rel),
			Folder:   filepath.ToSlash(folder),
			Ext:      ext,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Hash:     hash,
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

func deleteVideoByPath(root string, relPath string) error {
	if strings.TrimSpace(relPath) == "" {
		return errEmptyVideoPath
	}
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if clean == "." || strings.HasPrefix(clean, "..") {
		return errForbiddenPath
	}

	rootAbs, _ := filepath.Abs(root)
	fileAbs, err := filepath.Abs(filepath.Join(root, clean))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(fileAbs, rootAbs+string(os.PathSeparator)) {
		return errForbiddenPath
	}

	info, err := os.Stat(fileAbs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errVideoNotFound
		}
		return err
	}
	if info.IsDir() {
		return errNotAFile
	}

	if err := os.Remove(fileAbs); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errVideoNotFound
		}
		return err
	}
	return nil
}
