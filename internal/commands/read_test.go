package commands

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Christoph-D/pebbles/internal/peb"
)

func TestReadCommand(t *testing.T) {
	_, s, cleanup := setupTestStore(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Test content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	savedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb %s not found", id)
	}

	if savedPeb.ID != id {
		t.Errorf("expected ID %s, got %s", id, savedPeb.ID)
	}
	if savedPeb.Title != "Test peb" {
		t.Errorf("expected title 'Test peb', got '%s'", savedPeb.Title)
	}
	if savedPeb.Content != "Test content" {
		t.Errorf("expected content 'Test content', got '%s'", savedPeb.Content)
	}

	encoder := json.NewEncoder(&bytes.Buffer{})
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(savedPeb); err != nil {
		t.Fatal(err)
	}
}

func TestReadCommandNotFound(t *testing.T) {
	_, s, cleanup := setupTestStore(t)
	defer cleanup()

	_, ok := s.Get("peb-nonexistent")
	if ok {
		t.Error("expected peb to not be found")
	}
}

func TestReadCommandWithBlockedBy(t *testing.T) {
	_, s, cleanup := setupTestStore(t)
	defer cleanup()

	blockingID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	blockingPeb := peb.New(blockingID, "Blocking peb", peb.TypeTask, peb.StatusNew, "Blocks another peb")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Dependent peb", peb.TypeTask, peb.StatusNew, "Depends on blocking peb")
	p.BlockedBy = []string{blockingID}

	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	savedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb %s not found", id)
	}

	if len(savedPeb.BlockedBy) != 1 {
		t.Errorf("expected 1 blocked-by, got %d", len(savedPeb.BlockedBy))
	}
	if savedPeb.BlockedBy[0] != blockingID {
		t.Errorf("expected blocked-by %s, got %s", blockingID, savedPeb.BlockedBy[0])
	}
}

func TestReadCommandJSONOutput(t *testing.T) {
	_, s, cleanup := setupTestStore(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "JSON test", peb.TypeBug, peb.StatusNew, "Test content for JSON")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	savedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb %s not found", id)
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(savedPeb); err != nil {
		t.Fatal(err)
	}

	var decoded peb.Peb
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if decoded.ID != id {
		t.Errorf("expected ID %s in JSON, got %s", id, decoded.ID)
	}
	if decoded.Title != "JSON test" {
		t.Errorf("expected title 'JSON test' in JSON, got '%s'", decoded.Title)
	}
	if decoded.Content != "Test content for JSON" {
		t.Errorf("expected content 'Test content for JSON' in JSON, got '%s'", decoded.Content)
	}
}

func TestReadCommandFileIntegrity(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	expectedContent := "Multi-line\ncontent\nwith\nnewlines"
	p := peb.New(id, "File integrity test", peb.TypeFeature, peb.StatusInProgress, expectedContent)
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	filename := peb.Filename(p)
	filePath := filepath.Join(pebblesDir, filename)

	filePeb, err := peb.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read peb from file: %v", err)
	}

	if filePeb.ID != id {
		t.Errorf("expected ID %s in file, got %s", id, filePeb.ID)
	}
	if filePeb.Content != expectedContent {
		t.Errorf("expected content %q in file, got %q", expectedContent, filePeb.Content)
	}
}

func TestReadCommandMultipleIDs(t *testing.T) {
	_, s, cleanup := setupTestStore(t)
	defer cleanup()

	id1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	id3, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p1 := peb.New(id1, "First peb", peb.TypeTask, peb.StatusNew, "Content 1")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	p2 := peb.New(id2, "Second peb", peb.TypeBug, peb.StatusInProgress, "Content 2")
	if err := s.Save(p2); err != nil {
		t.Fatal(err)
	}

	p3 := peb.New(id3, "Third peb", peb.TypeFeature, peb.StatusFixed, "Content 3")
	if err := s.Save(p3); err != nil {
		t.Fatal(err)
	}

	pebs := []string{id1, id2, id3}
	retrieved := make([]peb.Peb, 0, len(pebs))

	for _, id := range pebs {
		p, ok := s.Get(id)
		if !ok {
			t.Fatalf("peb %s not found", id)
		}
		retrieved = append(retrieved, *p)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 pebs, got %d", len(retrieved))
	}

	if retrieved[0].ID != id1 || retrieved[0].Title != "First peb" {
		t.Errorf("first peb mismatch")
	}
	if retrieved[1].ID != id2 || retrieved[1].Title != "Second peb" {
		t.Errorf("second peb mismatch")
	}
	if retrieved[2].ID != id3 || retrieved[2].Title != "Third peb" {
		t.Errorf("third peb mismatch")
	}

	encoder := json.NewEncoder(&bytes.Buffer{})
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(retrieved); err != nil {
		t.Fatal(err)
	}
}
