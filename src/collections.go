package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const collectionsFileName = "watch_collections.json"

type WatchCollection struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	VideoHashes []string `json:"video_hashes"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type collectionsFile struct {
	Collections []WatchCollection `json:"collections"`
}

type WatchCollectionsStore struct {
	mu          sync.RWMutex
	path        string
	collections []WatchCollection
}

func newWatchCollectionsStore() (*WatchCollectionsStore, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(cfgDir, "vidvault")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	store := &WatchCollectionsStore{
		path:        filepath.Join(dir, collectionsFileName),
		collections: []WatchCollection{},
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *WatchCollectionsStore) load() error {
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

	var parsed collectionsFile
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}

	seenIDs := map[string]bool{}
	for _, c := range parsed.Collections {
		id := strings.TrimSpace(c.ID)
		name := strings.TrimSpace(c.Name)
		if id == "" || name == "" || seenIDs[id] {
			continue
		}
		seenIDs[id] = true

		cleanHashes := make([]string, 0, len(c.VideoHashes))
		for _, hash := range c.VideoHashes {
			hash = strings.TrimSpace(hash)
			if hash == "" || slices.Contains(cleanHashes, hash) {
				continue
			}
			cleanHashes = append(cleanHashes, hash)
		}

		createdAt := strings.TrimSpace(c.CreatedAt)
		updatedAt := strings.TrimSpace(c.UpdatedAt)
		if createdAt == "" {
			createdAt = time.Now().UTC().Format(time.RFC3339)
		}
		if updatedAt == "" {
			updatedAt = createdAt
		}

		s.collections = append(s.collections, WatchCollection{
			ID:          id,
			Name:        name,
			VideoHashes: cleanHashes,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	return nil
}

func (s *WatchCollectionsStore) Snapshot() []WatchCollection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]WatchCollection, len(s.collections))
	for i := range s.collections {
		out[i] = cloneCollection(s.collections[i])
	}
	return out
}

func (s *WatchCollectionsStore) Create(name string) (WatchCollection, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return WatchCollection{}, errors.New("missing collection name")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range s.collections {
		if strings.EqualFold(c.Name, trimmed) {
			return WatchCollection{}, errors.New("collection already exists")
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	collection := WatchCollection{
		ID:          newCollectionID(),
		Name:        trimmed,
		VideoHashes: []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.collections = append(s.collections, collection)
	if err := s.persistLocked(); err != nil {
		return WatchCollection{}, err
	}
	return cloneCollection(collection), nil
}

func (s *WatchCollectionsStore) Rename(id, name string) error {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" || name == "" {
		return errors.New("missing collection id or name")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range s.collections {
		if c.ID != id && strings.EqualFold(c.Name, name) {
			return errors.New("collection already exists")
		}
	}
	for i := range s.collections {
		if s.collections[i].ID != id {
			continue
		}
		s.collections[i].Name = name
		s.collections[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		return s.persistLocked()
	}
	return errors.New("collection not found")
}

func (s *WatchCollectionsStore) Delete(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("missing collection id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	next := make([]WatchCollection, 0, len(s.collections))
	found := false
	for _, c := range s.collections {
		if c.ID == id {
			found = true
			continue
		}
		next = append(next, c)
	}
	if !found {
		return errors.New("collection not found")
	}
	s.collections = next
	return s.persistLocked()
}

func (s *WatchCollectionsStore) SetVideo(id, hash string, assigned bool) error {
	id = strings.TrimSpace(id)
	hash = strings.TrimSpace(hash)
	if id == "" || hash == "" {
		return errors.New("missing collection id or hash")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.collections {
		if s.collections[i].ID != id {
			continue
		}
		current := s.collections[i].VideoHashes
		has := slices.Contains(current, hash)
		if assigned && !has {
			s.collections[i].VideoHashes = append(current, hash)
			s.collections[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return s.persistLocked()
		}
		if !assigned && has {
			next := make([]string, 0, len(current)-1)
			for _, existing := range current {
				if existing != hash {
					next = append(next, existing)
				}
			}
			s.collections[i].VideoHashes = next
			s.collections[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return s.persistLocked()
		}
		return nil
	}
	return errors.New("collection not found")
}

func (s *WatchCollectionsStore) AddVideos(id string, hashes []string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("missing collection id")
	}

	cleanHashes := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		hash = strings.TrimSpace(hash)
		if hash == "" || slices.Contains(cleanHashes, hash) {
			continue
		}
		cleanHashes = append(cleanHashes, hash)
	}
	if len(cleanHashes) == 0 {
		return errors.New("missing hashes")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.collections {
		if s.collections[i].ID != id {
			continue
		}
		changed := false
		for _, hash := range cleanHashes {
			if slices.Contains(s.collections[i].VideoHashes, hash) {
				continue
			}
			s.collections[i].VideoHashes = append(s.collections[i].VideoHashes, hash)
			changed = true
		}
		if changed {
			s.collections[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return s.persistLocked()
		}
		return nil
	}
	return errors.New("collection not found")
}

func (s *WatchCollectionsStore) persistLocked() error {
	payloadStruct := collectionsFile{Collections: s.collections}
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

func cloneCollection(c WatchCollection) WatchCollection {
	out := c
	out.VideoHashes = make([]string, len(c.VideoHashes))
	copy(out.VideoHashes, c.VideoHashes)
	return out
}

func newCollectionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("col_%d", os.Getpid())
	}
	return "col_" + hex.EncodeToString(b[:])
}
