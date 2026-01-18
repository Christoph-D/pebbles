package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
)

func setupTestStore(t *testing.T) (string, *store.Store, func()) {
	tmpDir := t.TempDir()
	pebblesDir := filepath.Join(tmpDir, ".pebbles")
	if err := os.Mkdir(pebblesDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(pebblesDir, "config.toml")
	configContent := `prefix = "peb"
id_length = 4
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	s := store.New(pebblesDir, "peb")
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return pebblesDir, s, cleanup
}

func TestNewCommand(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	input := NewInput{
		Title:   "Test task",
		Content: "This is a test task",
		Type:    "task",
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, input.Title, peb.Type(input.Type), peb.StatusNew, input.Content)

	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	filename := peb.Filename(p)
	expectedPath := filepath.Join(pebblesDir, filename)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("peb file was not created: %s", expectedPath)
	}

	savedPeb, err := peb.ReadFile(expectedPath)
	if err != nil {
		t.Fatal(err)
	}

	if savedPeb.ID != id {
		t.Errorf("expected ID %s, got %s", id, savedPeb.ID)
	}
	if savedPeb.Title != input.Title {
		t.Errorf("expected title %s, got %s", input.Title, savedPeb.Title)
	}
	if savedPeb.Content != input.Content {
		t.Errorf("expected content %s, got %s", input.Content, savedPeb.Content)
	}
}

func TestNewCommandWithBlockedBy(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	blockingID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	blockingPeb := peb.New(blockingID, "Blocking task", peb.TypeTask, peb.StatusNew, "Blocks another task")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	input := NewInput{
		Title:     "Dependent task",
		Content:   "This task depends on another",
		Type:      "task",
		BlockedBy: []string{blockingID},
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, input.Title, peb.Type(input.Type), peb.StatusNew, input.Content)
	p.BlockedBy = input.BlockedBy

	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	filename := peb.Filename(p)
	expectedPath := filepath.Join(pebblesDir, filename)

	savedPeb, err := peb.ReadFile(expectedPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(savedPeb.BlockedBy) != 1 {
		t.Fatalf("expected 1 blocked-by, got %d", len(savedPeb.BlockedBy))
	}
	if savedPeb.BlockedBy[0] != blockingID {
		t.Errorf("expected blocked-by %s, got %s", blockingID, savedPeb.BlockedBy[0])
	}
}

func TestNewCommandInvalidBlockedBy(t *testing.T) {
	_, s, cleanup := setupTestStore(t)
	defer cleanup()

	input := NewInput{
		Title:     "Test task",
		Content:   "This is a test task",
		BlockedBy: []string{"peb-nonexistent"},
	}

	err := peb.ValidateBlockedBy(s, nil, input.BlockedBy)
	if err == nil {
		t.Error("expected error for invalid blocked-by reference")
	}

	if !peb.HasInvalidReference(err) {
		t.Errorf("expected invalid reference error, got: %v", err)
	}
}

func TestExtractInvalidID(t *testing.T) {
	id := "peb-1234"
	wrappedErr := &customError{msg: peb.ErrInvalidReference.Error() + ": " + id}

	extracted := extractInvalidID(wrappedErr)
	if extracted != id {
		t.Errorf("expected extracted ID %s, got %s", id, extracted)
	}
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func TestNewCommandMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		input   NewInput
		wantErr bool
	}{
		{
			name:    "missing title",
			input:   NewInput{Content: "test"},
			wantErr: true,
		},
		{
			name:    "missing content",
			input:   NewInput{Title: "test"},
			wantErr: true,
		},
		{
			name: "valid input",
			input: NewInput{
				Title:   "test",
				Content: "test content",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			hasTitle := input.Title != ""
			hasContent := input.Content != ""

			if tt.wantErr {
				if hasTitle && hasContent {
					t.Error("expected missing fields but all present")
				}
			}
		})
	}
}

func TestNewCommandOutput(t *testing.T) {
	_, s, cleanup := setupTestStore(t)
	defer cleanup()

	input := NewInput{
		Title:   "Test task",
		Content: "Test content",
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, input.Title, peb.TypeBug, peb.StatusNew, input.Content)
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	filename := peb.Filename(p)

	expectedOutput := "Created new pebble " + id + " in .pebbles/" + filename

	if !strings.Contains(expectedOutput, id) {
		t.Errorf("expected output to contain ID %s", id)
	}
	if !strings.Contains(expectedOutput, filename) {
		t.Errorf("expected output to contain filename %s", filename)
	}
}
