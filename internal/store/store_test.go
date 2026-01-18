package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Christoph-D/pebbles/internal/peb"
)

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Test content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.Get(id); !ok {
		t.Error("expected peb to exist in store before deletion")
	}

	filename := peb.Filename(p)
	filePath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected peb file to exist before deletion")
	}

	if err := s.Delete(p); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	if _, ok := s.Get(id); ok {
		t.Error("expected peb to be removed from store map after deletion")
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected peb file to be deleted")
	}
}

func TestDeleteNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id := "peb-1234"
	p := peb.New(id, "Non-existent peb", peb.TypeTask, peb.StatusNew, "Content")

	if err := s.Delete(p); err == nil {
		t.Error("expected error when deleting non-existent file")
	}
}

func TestDeleteMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	var ids []string
	for i := 0; i < 3; i++ {
		id, err := s.GenerateUniqueID("peb", 4)
		if err != nil {
			t.Fatal(err)
		}
		p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
		if err := s.Save(p); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)

		if err := s.Delete(p); err != nil {
			t.Fatalf("Delete() failed for peb %s: %v", id, err)
		}
	}

	for _, id := range ids {
		if _, ok := s.Get(id); ok {
			t.Errorf("expected peb %s to be deleted", id)
		}
	}

	if len(s.All()) != 0 {
		t.Error("expected all pebs to be deleted")
	}
}

func TestDeleteAndReload(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p1 := peb.New(id1, "Peb 1", peb.TypeTask, peb.StatusNew, "Content 1")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p2 := peb.New(id2, "Peb 2", peb.TypeTask, peb.StatusNew, "Content 2")
	if err := s.Save(p2); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(p1); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	sReload := New(tmpDir)
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(id1); ok {
		t.Error("expected deleted peb to not exist after reload")
	}

	if _, ok := sReload.Get(id2); !ok {
		t.Error("expected non-deleted peb to exist after reload")
	}

	if len(sReload.All()) != 1 {
		t.Errorf("expected 1 peb after reload, got %d", len(sReload.All()))
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	id, err := peb.GenerateID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
	if err := peb.WriteFile(tmpDir, p); err != nil {
		t.Fatal(err)
	}

	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(s.All()) != 1 {
		t.Errorf("expected 1 peb after load, got %d", len(s.All()))
	}

	loadedPeb, ok := s.Get(id)
	if !ok {
		t.Error("expected peb to be loaded")
	}

	if loadedPeb.Title != p.Title {
		t.Errorf("expected title %q, got %q", p.Title, loadedPeb.Title)
	}
}

func TestLoadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load() failed on empty directory: %v", err)
	}

	if len(s.All()) != 0 {
		t.Errorf("expected 0 pebs, got %d", len(s.All()))
	}
}

func TestLoadNonExistentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "does-not-exist")

	s := New(nonExistentDir)
	if err := s.Load(); err == nil {
		t.Error("expected error when loading non-existent directory")
	}
}

func TestLoadIgnoresNonMarkdownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(s.All()) != 0 {
		t.Errorf("expected 0 pebs, got %d", len(s.All()))
	}
}

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	gotPeb, ok := s.Get(id)
	if !ok {
		t.Fatal("expected peb to be found")
	}

	if gotPeb.ID != id {
		t.Errorf("expected ID %s, got %s", id, gotPeb.ID)
	}

	if gotPeb.Title != p.Title {
		t.Errorf("expected title %q, got %q", p.Title, gotPeb.Title)
	}
}

func TestGetNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	_, ok := s.Get("non-existent-id")
	if ok {
		t.Error("expected peb not to be found")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")

	if err := s.Save(p); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	if _, ok := s.Get(id); !ok {
		t.Error("expected peb to be saved to store")
	}

	filename := peb.Filename(p)
	filePath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected peb file to be created")
	}
}

func TestSaveUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p := peb.New(id, "Original title", peb.TypeTask, peb.StatusNew, "Content")

	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	p.Title = "Updated title"
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatal("expected peb to be found")
	}

	if updatedPeb.Title != "Updated title" {
		t.Errorf("expected updated title, got %q", updatedPeb.Title)
	}
}

func TestAll(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	pebs := s.All()
	if len(pebs) != 0 {
		t.Errorf("expected 0 pebs, got %d", len(pebs))
	}

	for i := 0; i < 5; i++ {
		id, err := s.GenerateUniqueID("peb", 4)
		if err != nil {
			t.Fatal(err)
		}
		p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
		if err := s.Save(p); err != nil {
			t.Fatal(err)
		}
	}

	pebs = s.All()
	if len(pebs) != 5 {
		t.Errorf("expected 5 pebs, got %d", len(pebs))
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	if !s.Exists(id) {
		t.Error("expected peb to exist")
	}

	if s.Exists("non-existent-id") {
		t.Error("expected peb not to exist")
	}
}

func TestGenerateUniqueID(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatalf("GenerateUniqueID() failed: %v", err)
	}

	if len(id1) != 8 {
		t.Errorf("expected ID length 8 (peb- + 4 chars), got %d", len(id1))
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatalf("GenerateUniqueID() failed: %v", err)
	}

	if id1 == id2 {
		t.Error("expected unique IDs, got duplicates")
	}
}

func TestGenerateUniqueIDCollisionHandling(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	id := "peb-1234"
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	newID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatalf("GenerateUniqueID() failed: %v", err)
	}

	if newID == id {
		t.Error("expected unique ID different from existing")
	}
}

func TestGenerateUniqueIDFailure(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		id, err := s.GenerateUniqueID("peb", 1)
		if err != nil {
			return
		}
		p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
		if err := s.Save(p); err != nil {
			t.Fatal(err)
		}
	}
	t.Error("expected error when ID space is exhausted")
}
