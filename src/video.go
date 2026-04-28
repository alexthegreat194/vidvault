package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var videoLog = fileLogger("video")

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
	Favorite bool      `json:"is_favorite"`
	Tags     []string  `json:"tags,omitempty"` // list of tag IDs applied to this video hash
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
	logDebug(videoLog, "hashing file", "path", path)
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
	sum := hex.EncodeToString(hash.Sum(nil))
	logDebug(videoLog, "hashed file", "path", path, "hash", sum)
	return sum, nil
}

type videoCacheEntry struct {
	cacheKey string
	video    Video
}

type videoScanCache struct {
	mu      sync.RWMutex
	entries map[string]videoCacheEntry
}

func newVideoScanCache() *videoScanCache {
	return &videoScanCache{
		entries: map[string]videoCacheEntry{},
	}
}

func makeVideoCacheKey(info os.FileInfo) string {
	return strings.Join(
		[]string{
			strconv.FormatInt(info.Size(), 10),
			strconv.FormatInt(info.ModTime().UnixNano(), 10),
		},
		":",
	)
}

func (c *videoScanCache) getVideo(relPath string, cacheKey string) (Video, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[relPath]
	if !ok || entry.cacheKey != cacheKey {
		return Video{}, false
	}
	return entry.video, true
}

func (c *videoScanCache) setVideo(relPath string, cacheKey string, video Video) {
	c.mu.Lock()
	c.entries[relPath] = videoCacheEntry{
		cacheKey: cacheKey,
		video:    video,
	}
	c.mu.Unlock()
}

func (c *videoScanCache) prune(active map[string]struct{}) {
	c.mu.Lock()
	for relPath := range c.entries {
		if _, ok := active[relPath]; !ok {
			delete(c.entries, relPath)
		}
	}
	c.mu.Unlock()
}

func (c *videoScanCache) size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func normalizeVideoFromRelPath(video Video, relPath string) Video {
	video.Path = filepath.ToSlash(relPath)
	video.Name = filepath.Base(relPath)
	if video.Folder == "" {
		folder := filepath.Dir(relPath)
		if folder == "." {
			folder = "/"
		}
		video.Folder = filepath.ToSlash(folder)
	}
	return video
}

func buildVideo(root string, relPath string, ext string, info os.FileInfo) (Video, error) {
	absPath := filepath.Join(root, relPath)
	hash, err := hashFile(absPath)
	if err != nil {
		return Video{}, err
	}
	folder := filepath.Dir(relPath)
	if folder == "." {
		folder = "/"
	}
	return Video{
		Name:     filepath.Base(relPath),
		Path:     filepath.ToSlash(relPath),
		Folder:   filepath.ToSlash(folder),
		Ext:      ext,
		Size:     info.Size(),
		Modified: info.ModTime(),
		Hash:     hash,
	}, nil
}

func scanVideosWithCallback(root string, cache *videoScanCache, onVideo func(Video) error) ([]Video, error) {
	logDebug(videoLog, "starting video scan", "root", root)
	var videos []Video
	activePaths := map[string]struct{}{}
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
		rel = filepath.Clean(rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		cacheKey := makeVideoCacheKey(info)
		activePaths[rel] = struct{}{}

		var video Video
		if cache != nil {
			if cached, ok := cache.getVideo(rel, cacheKey); ok {
				video = normalizeVideoFromRelPath(cached, rel)
			}
		}
		if video.Hash == "" {
			video, err = buildVideo(root, rel, ext, info)
			if err != nil {
				return err
			}
			if cache != nil {
				cache.setVideo(rel, cacheKey, video)
			}
		} else {
			video.Ext = ext
			video.Size = info.Size()
			video.Modified = info.ModTime()
		}

		videos = append(videos, video)
		if onVideo != nil {
			if err := onVideo(video); err != nil {
				return err
			}
		}
		logDebug(videoLog, "video discovered", "path", video.Path, "size", info.Size(), "ext", ext)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Path < videos[j].Path
	})
	if cache != nil {
		cache.prune(activePaths)
		logDebug(videoLog, "video scan cache pruned", "cache_entries", cache.size())
	}
	videoLog.Info("video scan completed", "root", root, "count", len(videos))
	return videos, nil
}

// Walks root recursively and returns all video files found, sorted
// by their relative path. Files with extensions not in videoExts are skipped.
func scanVideos(root string, cache *videoScanCache) ([]Video, error) {
	return scanVideosWithCallback(root, cache, nil)
}

func deleteVideoByPath(root string, relPath string) error {
	logDebug(videoLog, "delete requested", "root", root, "path", relPath)
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
	videoLog.Info("video deleted", "path", relPath, "abs_path", fileAbs)
	return nil
}
