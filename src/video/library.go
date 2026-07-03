package video

import (
	"sort"
	"sync"
)

// VideoLibrary is an in-memory index of discovered videos keyed by relative path.
type VideoLibrary struct {
	mu     sync.RWMutex
	videos map[string]Video
	ready  bool
}

// NewVideoLibrary creates an empty VideoLibrary.
func NewVideoLibrary() *VideoLibrary {
	return &VideoLibrary{
		videos: map[string]Video{},
	}
}

// Upsert inserts or replaces a video by path.
func (l *VideoLibrary) Upsert(v Video) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.videos[v.Path] = v
}

// Remove deletes a video by path and returns the removed entry if present.
func (l *VideoLibrary) Remove(path string) (Video, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.videos[path]
	if ok {
		delete(l.videos, path)
	}
	return v, ok
}

// Get returns a video by path.
func (l *VideoLibrary) Get(path string) (Video, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	v, ok := l.videos[path]
	return v, ok
}

// AllSorted returns all videos sorted by path.
func (l *VideoLibrary) AllSorted() []Video {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Video, 0, len(l.videos))
	for _, v := range l.videos {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

// ReplaceAll replaces the entire index with the given videos.
func (l *VideoLibrary) ReplaceAll(videos []Video) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.videos = make(map[string]Video, len(videos))
	for _, v := range videos {
		l.videos[v.Path] = v
	}
}

// IsReady reports whether the library has completed its initial scan.
func (l *VideoLibrary) IsReady() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.ready
}

// SetReady sets the ready flag.
func (l *VideoLibrary) SetReady(ready bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ready = ready
}

// Count returns the number of indexed videos.
func (l *VideoLibrary) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.videos)
}
