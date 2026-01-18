package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Christoph-D/pebbles/internal/peb"
)

type Store struct {
	pebs map[string]*peb.Peb
	dir  string
}

func New(dir string) *Store {
	return &Store{
		pebs: make(map[string]*peb.Peb),
		dir:  dir,
	}
}

func (s *Store) Load() error {
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

		path := filepath.Join(s.dir, name)
		p, err := peb.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", name, err)
		}

		s.pebs[p.ID] = p
	}

	return nil
}

func (s *Store) Get(id string) (*peb.Peb, bool) {
	p, ok := s.pebs[id]
	return p, ok
}

func (s *Store) Save(p *peb.Peb) error {
	if err := peb.WriteFile(s.dir, p); err != nil {
		return fmt.Errorf("failed to save peb: %w", err)
	}
	s.pebs[p.ID] = p
	return nil
}

func (s *Store) All() []*peb.Peb {
	result := make([]*peb.Peb, 0, len(s.pebs))
	for _, p := range s.pebs {
		result = append(result, p)
	}
	return result
}

func (s *Store) Exists(id string) bool {
	_, ok := s.pebs[id]
	return ok
}

func (s *Store) Delete(p *peb.Peb) error {
	filename := peb.Filename(p)
	path := filepath.Join(s.dir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete peb file: %w", err)
	}
	delete(s.pebs, p.ID)
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
