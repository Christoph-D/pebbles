package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Christoph-D/pebbles/internal/peb"
)

type Store struct {
	cache     map[string]*peb.Peb
	filenames map[string]string
	dir       string
	prefix    string
}

func New(dir string, prefix string) *Store {
	return &Store{
		cache:     make(map[string]*peb.Peb),
		filenames: make(map[string]string),
		dir:       dir,
		prefix:    prefix,
	}
}

func (s *Store) Load() error {
	s.cache = make(map[string]*peb.Peb)
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("failed to read pebbles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		id, err := peb.ParseID(name, s.prefix)
		if err != nil {
			continue
		}
		s.filenames[id] = name
	}

	return nil
}

func (s *Store) Get(id string) (*peb.Peb, bool) {
	if p, ok := s.cache[id]; ok {
		return p, true
	}

	filename, ok := s.filenames[id]
	if !ok {
		return nil, false
	}

	path := filepath.Join(s.dir, filename)
	p, err := peb.ReadFile(path)
	if err != nil {
		return nil, false
	}

	filtered := make([]string, 0, len(p.BlockedBy))
	for _, bid := range p.BlockedBy {
		if s.Exists(bid) {
			filtered = append(filtered, bid)
		}
	}
	p.BlockedBy = filtered

	s.cache[id] = p
	return p, true
}

func (s *Store) Save(p *peb.Peb) error {
	cleaned := *p
	filtered := make([]string, 0, len(p.BlockedBy))
	for _, id := range p.BlockedBy {
		if s.Exists(id) {
			filtered = append(filtered, id)
		}
	}
	cleaned.BlockedBy = filtered

	if err := peb.WriteFile(s.dir, &cleaned); err != nil {
		return fmt.Errorf("failed to save peb: %w", err)
	}
	s.cache[cleaned.ID] = &cleaned
	s.filenames[cleaned.ID] = peb.Filename(&cleaned)
	return nil
}

func (s *Store) All() []*peb.Peb {
	result := make([]*peb.Peb, 0, len(s.filenames))
	for id := range s.filenames {
		if p, ok := s.Get(id); ok {
			result = append(result, p)
		}
	}
	return result
}

func (s *Store) Exists(id string) bool {
	_, ok := s.filenames[id]
	return ok
}

func (s *Store) Delete(p *peb.Peb) error {
	filename := peb.Filename(p)
	path := filepath.Join(s.dir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete peb file: %w", err)
	}
	delete(s.cache, p.ID)
	delete(s.filenames, p.ID)
	return nil
}

func (s *Store) GenerateUniqueID(prefix string, length int) (string, error) {
	const maxAttempts = 1000
	for range maxAttempts {
		id, err := peb.GenerateID(prefix, length)
		if err != nil {
			return "", err
		}
		if !s.Exists(id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique ID after %d attempts", maxAttempts)
}
