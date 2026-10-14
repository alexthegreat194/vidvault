package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const uploadBatchesFileName = "upload_batches.json"

const (
	batchSourceUpload     = "upload"
	batchSourceMigrated   = "migrated"
	batchSourceDiscovered = "discovered"
)

type UploadBatch struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CreatedAt    string `json:"created_at"`
	NeedsSorting bool   `json:"needs_sorting"`
	Source       string `json:"source"`
}

type UploadBatchResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CreatedAt    string `json:"created_at"`
	NeedsSorting bool   `json:"needs_sorting"`
	Source       string `json:"source"`
	VideoCount   int    `json:"video_count"`
	NewCount     int    `json:"new_count"`
}

type uploadBatchesDiskState struct {
	BatchesMigrated bool          `json:"batches_migrated"`
	Batches         []UploadBatch `json:"batches"`
}

type UploadBatchesStore struct {
	mu              sync.RWMutex
	path            string
	batchesMigrated bool
	batches         []UploadBatch
}

func newUploadBatchesStore(seen *SeenFilesStore) (*UploadBatchesStore, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(cfgDir, "vidvault")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	store := &UploadBatchesStore{
		path:    filepath.Join(dir, uploadBatchesFileName),
		batches: []UploadBatch{},
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	if err := store.MigrateSeenFiles(seen); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *UploadBatchesStore) Get(id string) (UploadBatch, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return UploadBatch{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.batches {
		if b.ID == id {
			return b, true
		}
	}
	return UploadBatch{}, false
}

func (s *UploadBatchesStore) AssignDiscoveredBatch() (string, error) {
	return s.GetOrCreateDiscoveredBatch()
}

func (s *UploadBatchesStore) ListWithStats(seen *SeenFilesStore) []UploadBatchResponse {
	return s.ListWithCounts(seen)
}

func (s *UploadBatchesStore) load() error {
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
	var parsed uploadBatchesDiskState
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	s.batchesMigrated = parsed.BatchesMigrated
	if parsed.Batches != nil {
		s.batches = parsed.Batches
	}
	return nil
}

func defaultBatchName(t time.Time) string {
	return t.Local().Format("Jan 2, 3:04 PM")
}

func newBatchID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("batch_%d", time.Now().UnixNano())
	}
	return "batch_" + hex.EncodeToString(b[:])
}

func parseBatchTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

func hourWindowKey(firstSeenAt string) string {
	t := parseBatchTime(firstSeenAt)
	if t.IsZero() {
		return strings.TrimSpace(firstSeenAt)
	}
	return t.Truncate(time.Hour).UTC().Format(time.RFC3339)
}

func (s *UploadBatchesStore) Exists(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.batches {
		if b.ID == id {
			return true
		}
	}
	return false
}

func (s *UploadBatchesStore) Create(source string) (UploadBatch, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		source = batchSourceUpload
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createLocked(source, time.Now().UTC())
}

func (s *UploadBatchesStore) createLocked(source string, at time.Time) (UploadBatch, error) {
	createdAt := at.UTC().Format(time.RFC3339)
	batch := UploadBatch{
		ID:           newBatchID(),
		Name:         defaultBatchName(at),
		CreatedAt:    createdAt,
		NeedsSorting: true,
		Source:       source,
	}
	s.batches = append(s.batches, batch)
	if err := s.persistLocked(); err != nil {
		return UploadBatch{}, err
	}
	return batch, nil
}

func (s *UploadBatchesStore) GetOrCreateDiscoveredBatch() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().UTC().Add(-30 * time.Minute)
	for i := len(s.batches) - 1; i >= 0; i-- {
		b := s.batches[i]
		if b.Source != batchSourceDiscovered {
			continue
		}
		created := parseBatchTime(b.CreatedAt)
		if !created.IsZero() && !created.Before(cutoff) {
			return b.ID, nil
		}
	}
	batch, err := s.createLocked(batchSourceDiscovered, time.Now().UTC())
	if err != nil {
		return "", err
	}
	return batch.ID, nil
}

func (s *UploadBatchesStore) Update(id string, name *string, needsSorting *bool) (UploadBatch, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return UploadBatch{}, errors.New("missing batch id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.batches {
		if s.batches[i].ID != id {
			continue
		}
		if name != nil {
			trimmed := strings.TrimSpace(*name)
			if trimmed == "" {
				created := parseBatchTime(s.batches[i].CreatedAt)
				if created.IsZero() {
					created = time.Now().UTC()
				}
				s.batches[i].Name = defaultBatchName(created)
			} else {
				s.batches[i].Name = trimmed
			}
		}
		if needsSorting != nil {
			s.batches[i].NeedsSorting = *needsSorting
		}
		if err := s.persistLocked(); err != nil {
			return UploadBatch{}, err
		}
		return s.batches[i], nil
	}
	return UploadBatch{}, errors.New("batch not found")
}

func (s *UploadBatchesStore) ListWithCounts(seen *SeenFilesStore) []UploadBatchResponse {
	counts, newCounts := seen.batchCounts()

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]UploadBatchResponse, 0, len(s.batches))
	for _, b := range s.batches {
		out = append(out, UploadBatchResponse{
			ID:           b.ID,
			Name:         b.Name,
			CreatedAt:    b.CreatedAt,
			NeedsSorting: b.NeedsSorting,
			Source:       b.Source,
			VideoCount:   counts[b.ID],
			NewCount:     newCounts[b.ID],
		})
	}

	sort.Slice(out, func(i, j int) bool {
		ti := parseBatchTime(out[i].CreatedAt)
		tj := parseBatchTime(out[j].CreatedAt)
		if ti.Equal(tj) {
			return out[i].ID > out[j].ID
		}
		return ti.After(tj)
	})
	return out
}

func (s *UploadBatchesStore) MigrateSeenFiles(seen *SeenFilesStore) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.batchesMigrated {
		return nil
	}

	seen.mu.Lock()
	defer seen.mu.Unlock()

	type groupEntry struct {
		key    string
		record seenFileRecord
	}
	groups := map[string][]groupEntry{}

	for key, record := range seen.entries {
		if strings.TrimSpace(record.BatchID) != "" {
			continue
		}
		window := hourWindowKey(record.FirstSeenAt)
		groups[window] = append(groups[window], groupEntry{key: key, record: record})
	}

	windowKeys := make([]string, 0, len(groups))
	for window := range groups {
		windowKeys = append(windowKeys, window)
	}
	sort.Strings(windowKeys)

	for _, window := range windowKeys {
		entries := groups[window]
		if len(entries) == 0 {
			continue
		}

		needsSorting := false
		createdAt := time.Now().UTC()
		for _, entry := range entries {
			if strings.TrimSpace(entry.record.ReviewedAt) == "" {
				needsSorting = true
			}
			if t := parseBatchTime(entry.record.FirstSeenAt); !t.IsZero() && (createdAt.Equal(time.Now().UTC()) || t.Before(createdAt)) {
				createdAt = t
			}
		}
		if createdAt.Equal(time.Now().UTC()) {
			if t := parseBatchTime(window); !t.IsZero() {
				createdAt = t
			}
		}

		batch := UploadBatch{
			ID:           newBatchID(),
			Name:         defaultBatchName(createdAt),
			CreatedAt:    createdAt.UTC().Format(time.RFC3339),
			NeedsSorting: needsSorting,
			Source:       batchSourceMigrated,
		}
		s.batches = append(s.batches, batch)

		for _, entry := range entries {
			record := entry.record
			record.BatchID = batch.ID
			seen.entries[entry.key] = record
		}
	}

	s.batchesMigrated = true
	if err := s.persistLocked(); err != nil {
		return err
	}
	return seen.persistLocked()
}

func (s *UploadBatchesStore) persistLocked() error {
	payloadStruct := uploadBatchesDiskState{
		BatchesMigrated: s.batchesMigrated,
		Batches:         s.batches,
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
