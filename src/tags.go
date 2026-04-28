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
)

const tagsFileName = "tags.json"

type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type tagsFile struct {
	Tags        []Tag               `json:"tags"`
	Assignments map[string][]string `json:"assignments"`
}

type TagsStore struct {
	mu          sync.RWMutex
	path        string
	tags        []Tag
	assignments map[string][]string
}

func newTagsStore() (*TagsStore, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(cfgDir, "vidvault")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	store := &TagsStore{
		path:        filepath.Join(dir, tagsFileName),
		tags:        []Tag{},
		assignments: map[string][]string{},
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *TagsStore) load() error {
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

	var parsed tagsFile
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}

	seenIDs := map[string]bool{}
	for _, tag := range parsed.Tags {
		if strings.TrimSpace(tag.ID) == "" || strings.TrimSpace(tag.Name) == "" {
			continue
		}
		if seenIDs[tag.ID] {
			continue
		}
		if strings.TrimSpace(tag.Color) == "" {
			tag.Color = randomTagColor()
		}
		seenIDs[tag.ID] = true
		s.tags = append(s.tags, tag)
	}

	for hash, ids := range parsed.Assignments {
		if strings.TrimSpace(hash) == "" {
			continue
		}
		clean := make([]string, 0, len(ids))
		for _, id := range ids {
			if !seenIDs[id] {
				continue
			}
			if slices.Contains(clean, id) {
				continue
			}
			clean = append(clean, id)
		}
		if len(clean) > 0 {
			s.assignments[hash] = clean
		}
	}
	return nil
}

func (s *TagsStore) Snapshot() tagsFile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tags := make([]Tag, len(s.tags))
	copy(tags, s.tags)

	assignments := make(map[string][]string, len(s.assignments))
	for hash, ids := range s.assignments {
		out := make([]string, len(ids))
		copy(out, ids)
		assignments[hash] = out
	}

	return tagsFile{
		Tags:        tags,
		Assignments: assignments,
	}
}

func (s *TagsStore) TagsForHash(hash string) []string {
	if strings.TrimSpace(hash) == "" {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.assignments[hash]
	out := make([]string, len(ids))
	copy(out, ids)
	return out
}

func (s *TagsStore) Create(name string) (Tag, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return Tag{}, errors.New("missing tag name")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.tags {
		if strings.EqualFold(t.Name, trimmed) {
			return Tag{}, errors.New("tag already exists")
		}
	}

	tag := Tag{
		ID:    newTagID(),
		Name:  trimmed,
		Color: randomTagColor(),
	}
	s.tags = append(s.tags, tag)
	if err := s.persistLocked(); err != nil {
		return Tag{}, err
	}
	return tag, nil
}

func (s *TagsStore) Delete(tagID string) error {
	tagID = strings.TrimSpace(tagID)
	if tagID == "" {
		return errors.New("missing tag id")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	nextTags := make([]Tag, 0, len(s.tags))
	found := false
	for _, t := range s.tags {
		if t.ID == tagID {
			found = true
			continue
		}
		nextTags = append(nextTags, t)
	}
	if !found {
		return errors.New("tag not found")
	}
	s.tags = nextTags

	for hash, ids := range s.assignments {
		nextIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != tagID {
				nextIDs = append(nextIDs, id)
			}
		}
		if len(nextIDs) == 0 {
			delete(s.assignments, hash)
			continue
		}
		s.assignments[hash] = nextIDs
	}

	return s.persistLocked()
}

func (s *TagsStore) SetAssignment(hash, tagID string, assigned bool) error {
	hash = strings.TrimSpace(hash)
	tagID = strings.TrimSpace(tagID)
	if hash == "" || tagID == "" {
		return errors.New("missing hash or tag id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.hasTagLocked(tagID) {
		return errors.New("tag not found")
	}

	current := s.assignments[hash]
	has := slices.Contains(current, tagID)

	if assigned && !has {
		s.assignments[hash] = append(current, tagID)
	}
	if !assigned && has {
		next := make([]string, 0, len(current)-1)
		for _, id := range current {
			if id != tagID {
				next = append(next, id)
			}
		}
		if len(next) == 0 {
			delete(s.assignments, hash)
		} else {
			s.assignments[hash] = next
		}
	}

	return s.persistLocked()
}

func (s *TagsStore) hasTagLocked(tagID string) bool {
	for _, t := range s.tags {
		if t.ID == tagID {
			return true
		}
	}
	return false
}

func (s *TagsStore) persistLocked() error {
	payloadStruct := tagsFile{
		Tags:        s.tags,
		Assignments: s.assignments,
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

func newTagID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("tag_%d", os.Getpid())
	}
	return "tag_" + hex.EncodeToString(b[:])
}

func randomTagColor() string {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "#888888"
	}

	// Keep colors visible on dark backgrounds by biasing to mid/high channels.
	r := 80 + int(b[0])%136
	g := 80 + int(b[1])%136
	bl := 80 + int(b[2])%136
	return fmt.Sprintf("#%02x%02x%02x", r, g, bl)
}
