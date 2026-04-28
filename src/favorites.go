package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const favoritesFileName = "favorites.json"

var favoritesLog = fileLogger("favorites")

// FavoritesStore persists favorite state keyed by video hash.
type FavoritesStore struct {
	mu   sync.RWMutex
	path string
	data map[string]bool
}

func newFavoritesStore() (*FavoritesStore, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		favoritesLog.Error("failed to resolve user config dir", "error", err)
		return nil, err
	}
	dir := filepath.Join(cfgDir, "vidvault")
	if err := os.MkdirAll(dir, 0755); err != nil {
		favoritesLog.Error("failed to ensure favorites config dir", "dir", dir, "error", err)
		return nil, err
	}

	store := &FavoritesStore{
		path: filepath.Join(dir, favoritesFileName),
		data: map[string]bool{},
	}
	if err := store.load(); err != nil {
		favoritesLog.Error("failed loading favorites store", "path", store.path, "error", err)
		return nil, err
	}
	logDebug(favoritesLog, "favorites store initialized", "path", store.path, "entries", len(store.data))
	return store, nil
}

func (s *FavoritesStore) load() error {
	logDebug(favoritesLog, "loading favorites", "path", s.path)
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logDebug(favoritesLog, "favorites file does not exist yet", "path", s.path)
			return nil
		}
		return err
	}
	if len(raw) == 0 {
		return nil
	}

	var parsed map[string]bool
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	for hash, isFavorite := range parsed {
		if hash == "" || !isFavorite {
			continue
		}
		s.data[hash] = true
	}
	logDebug(favoritesLog, "favorites loaded", "count", len(s.data))
	return nil
}

func (s *FavoritesStore) IsFavorite(hash string) bool {
	if hash == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[hash]
}

func (s *FavoritesStore) Snapshot() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]bool, len(s.data))
	for hash := range s.data {
		out[hash] = true
	}
	return out
}

func (s *FavoritesStore) Set(hash string, favorite bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	logDebug(favoritesLog, "updating favorite state", "hash", hash, "favorite", favorite)

	if favorite {
		s.data[hash] = true
	} else {
		delete(s.data, hash)
	}
	return s.persistLocked()
}

func (s *FavoritesStore) persistLocked() error {
	logDebug(favoritesLog, "persisting favorites", "path", s.path, "count", len(s.data))
	payload, err := json.MarshalIndent(s.data, "", "  ")
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
	logDebug(favoritesLog, "favorites persisted successfully", "path", s.path)
	return nil
}
