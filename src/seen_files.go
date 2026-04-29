package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const seenFilesFileName = "seen_files.json"

type seenFileRecord struct {
	FirstSeenAt string `json:"first_seen_at"`
	ReviewedAt  string `json:"reviewed_at,omitempty"`
}

type seenFilesDiskState struct {
	Initialized    bool                      `json:"initialized,omitempty"` // legacy
	BaselinedRoots map[string]bool           `json:"baselined_roots,omitempty"`
	Entries        map[string]seenFileRecord `json:"entries"`
}

type SeenFilesStore struct {
	mu            sync.RWMutex
	path          string
	rootBaselines map[string]bool
	entries       map[string]seenFileRecord
}

func newSeenFilesStore() (*SeenFilesStore, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(cfgDir, "vidvault")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	store := &SeenFilesStore{
		path:          filepath.Join(dir, seenFilesFileName),
		rootBaselines: map[string]bool{},
		entries:       map[string]seenFileRecord{},
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *SeenFilesStore) load() error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	var parsed seenFilesDiskState
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	if parsed.BaselinedRoots != nil {
		s.rootBaselines = parsed.BaselinedRoots
	}
	if parsed.Initialized && len(s.rootBaselines) == 0 {
		s.rootBaselines["__legacy__"] = true
	}
	if parsed.Entries != nil {
		s.entries = parsed.Entries
	}
	return nil
}

func normalizeRootKey(root string) string {
	if strings.TrimSpace(root) == "" {
		return "__unknown__"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	return filepath.ToSlash(filepath.Clean(abs))
}

func (s *SeenFilesStore) keyForVideo(v Video) string {
	if hash := strings.TrimSpace(v.Hash); hash != "" {
		return "hash:" + hash
	}
	cleanPath := filepath.ToSlash(filepath.Clean(v.Path))
	if cleanPath == "." || cleanPath == "" {
		return ""
	}
	return "path:" + cleanPath
}

// MarkVideos mutates each video's IsNew flag and persists seen-state changes.
func (s *SeenFilesStore) MarkVideos(root string, videos []Video) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rootKey := normalizeRootKey(root)
	now := time.Now().UTC().Format(time.RFC3339)
	changed := false

	if !s.rootBaselines[rootKey] && !s.rootBaselines["__legacy__"] {
		for i := range videos {
			key := s.keyForVideo(videos[i])
			if key == "" {
				continue
			}
			if _, ok := s.entries[key]; !ok {
				s.entries[key] = seenFileRecord{
					FirstSeenAt: now,
					ReviewedAt:  now,
				}
				changed = true
			}
			videos[i].IsNew = false
		}
		s.rootBaselines[rootKey] = true
		changed = true
		if changed {
			return s.persistLocked()
		}
		return nil
	}

	for i := range videos {
		key := s.keyForVideo(videos[i])
		if key == "" {
			continue
		}
		record, ok := s.entries[key]
		if !ok {
			s.entries[key] = seenFileRecord{FirstSeenAt: now}
			videos[i].IsNew = true
			changed = true
			continue
		}
		videos[i].IsNew = strings.TrimSpace(record.ReviewedAt) == ""
	}

	if changed {
		return s.persistLocked()
	}
	return nil
}

func (s *SeenFilesStore) MarkReviewedByKeys(keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	changed := false
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		record, ok := s.entries[key]
		if !ok {
			s.entries[key] = seenFileRecord{
				FirstSeenAt: now,
				ReviewedAt:  now,
			}
			changed = true
			continue
		}
		if strings.TrimSpace(record.ReviewedAt) != "" {
			continue
		}
		record.ReviewedAt = now
		s.entries[key] = record
		changed = true
	}
	if !changed {
		return nil
	}
	return s.persistLocked()
}

func (s *SeenFilesStore) MarkReviewedForVideos(videos []Video, clearAll bool, paths []string, hashes []string) error {
	pathSet := map[string]struct{}{}
	hashSet := map[string]struct{}{}
	for _, path := range paths {
		path = filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
		if path == "" || path == "." {
			continue
		}
		pathSet[path] = struct{}{}
	}
	for _, hash := range hashes {
		hash = strings.TrimSpace(hash)
		if hash == "" {
			continue
		}
		hashSet[hash] = struct{}{}
	}
	keys := make([]string, 0, len(videos))
	for _, video := range videos {
		if !video.IsNew {
			continue
		}
		if !clearAll {
			pathMatch := false
			hashMatch := false
			if _, ok := pathSet[filepath.ToSlash(filepath.Clean(video.Path))]; ok {
				pathMatch = true
			}
			if _, ok := hashSet[strings.TrimSpace(video.Hash)]; ok {
				hashMatch = true
			}
			if !pathMatch && !hashMatch {
				continue
			}
		}
		if key := s.keyForVideo(video); key != "" {
			keys = append(keys, key)
		}
	}
	return s.MarkReviewedByKeys(keys)
}

func (s *SeenFilesStore) ForgetForVideos(videos []Video, forgetAll bool, paths []string, hashes []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pathSet := map[string]struct{}{}
	hashSet := map[string]struct{}{}
	for _, path := range paths {
		path = filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
		if path == "" || path == "." {
			continue
		}
		pathSet[path] = struct{}{}
	}
	for _, hash := range hashes {
		hash = strings.TrimSpace(hash)
		if hash == "" {
			continue
		}
		hashSet[hash] = struct{}{}
	}

	changed := false
	for _, video := range videos {
		if !forgetAll {
			pathMatch := false
			hashMatch := false
			if _, ok := pathSet[filepath.ToSlash(filepath.Clean(video.Path))]; ok {
				pathMatch = true
			}
			if _, ok := hashSet[strings.TrimSpace(video.Hash)]; ok {
				hashMatch = true
			}
			if !pathMatch && !hashMatch {
				continue
			}
		}
		key := s.keyForVideo(video)
		if key == "" {
			continue
		}
		if _, ok := s.entries[key]; !ok {
			continue
		}
		delete(s.entries, key)
		changed = true
	}
	if !changed {
		return nil
	}
	return s.persistLocked()
}

func (s *SeenFilesStore) persistLocked() error {
	payloadStruct := seenFilesDiskState{
		BaselinedRoots: s.rootBaselines,
		Entries:        s.entries,
	}
	payload, err := json.MarshalIndent(payloadStruct, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
