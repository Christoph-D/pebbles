package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

func setupTestStoreForUpdate(t *testing.T) (string, *store.Store, func()) {
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

	s := store.New(pebblesDir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return pebblesDir, s, cleanup
}

func TestUpdateCommandStatus(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	update := UpdateInput{
		Status: stringPtr("in-progress"),
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if updatedPeb.Status != peb.StatusInProgress {
		t.Errorf("expected status in-progress, got %s", updatedPeb.Status)
	}

	if !strings.Contains(output, "Updated status of "+id+".") {
		t.Errorf("expected output to contain status update message, got: %s", output)
	}
}

func TestUpdateCommandTitle(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	oldTitle := "Test task"
	newTitle := "Updated task title"
	p := peb.New(id, oldTitle, peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	oldFilename := peb.Filename(p)
	oldPath := filepath.Join(pebblesDir, oldFilename)
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		t.Fatalf("original peb file not found: %s", oldPath)
	}

	update := UpdateInput{
		Title: &newTitle,
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if updatedPeb.Title != newTitle {
		t.Errorf("expected title %s, got %s", newTitle, updatedPeb.Title)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("old file still exists after rename: %s", oldPath)
	}

	newFilename := peb.Filename(updatedPeb)
	newPath := filepath.Join(pebblesDir, newFilename)
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("new file not found after rename: %s", newPath)
	}

	if !strings.Contains(output, "Updated title of "+id+".") {
		t.Errorf("expected output to contain title update message, got: %s", output)
	}
}

func TestUpdateCommandContent(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	newContent := "Updated content with more details"
	update := UpdateInput{
		Content: &newContent,
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if updatedPeb.Content != newContent {
		t.Errorf("expected content %s, got %s", newContent, updatedPeb.Content)
	}

	if !strings.Contains(output, "Updated content of "+id+".") {
		t.Errorf("expected output to contain content update message, got: %s", output)
	}
}

func TestUpdateCommandType(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	newType := "bug"
	update := UpdateInput{
		Type: &newType,
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if updatedPeb.Type != peb.TypeBug {
		t.Errorf("expected type bug, got %s", updatedPeb.Type)
	}

	if !strings.Contains(output, "Updated type of "+id+".") {
		t.Errorf("expected output to contain type update message, got: %s", output)
	}
}

func TestUpdateCommandBlockedBy(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	blockingID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	blockingPeb := peb.New(blockingID, "Blocking task", peb.TypeTask, peb.StatusNew, "Blocks another task")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	blockedBy := []string{blockingID}
	update := UpdateInput{
		BlockedBy: &blockedBy,
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if len(updatedPeb.BlockedBy) != 1 {
		t.Errorf("expected 1 blocked-by, got %d", len(updatedPeb.BlockedBy))
	}
	if updatedPeb.BlockedBy[0] != blockingID {
		t.Errorf("expected blocked-by %s, got %s", blockingID, updatedPeb.BlockedBy[0])
	}

	if !strings.Contains(output, "Updated blocked-by list of "+id+".") {
		t.Errorf("expected output to contain blocked-by update message, got: %s", output)
	}
}

func TestUpdateCommandClearBlockedBy(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	blockingID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	blockingPeb := peb.New(blockingID, "Blocking task", peb.TypeTask, peb.StatusNew, "Blocks another task")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	p.BlockedBy = []string{blockingID}
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	update := UpdateInput{
		BlockedBy: &[]string{},
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if len(updatedPeb.BlockedBy) != 0 {
		t.Errorf("expected 0 blocked-by, got %d", len(updatedPeb.BlockedBy))
	}

	if !strings.Contains(output, "Cleared blocked-by list of "+id+".") {
		t.Errorf("expected output to contain cleared blocked-by message, got: %s", output)
	}
}

func TestUpdateCommandInvalidBlockedBy(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	blockedBy := []string{"peb-nonexistent"}
	update := UpdateInput{
		BlockedBy: &blockedBy,
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if !strings.Contains(output, "Error:") {
		t.Error("expected error for invalid blocked-by reference")
	}

	if !strings.Contains(output, "referenced pebble(s) not found") {
		t.Errorf("expected invalid reference error, got: %s", output)
	}
}

func TestUpdateCommandCycleDetection(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p1 := peb.New(id1, "Task 1", peb.TypeTask, peb.StatusNew, "First task")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p2 := peb.New(id2, "Task 2", peb.TypeTask, peb.StatusNew, "Second task")
	p2.BlockedBy = []string{id1}
	if err := s.Save(p2); err != nil {
		t.Fatal(err)
	}

	blockedBy := []string{id2}
	update := UpdateInput{
		BlockedBy: &blockedBy,
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id1, string(inputJSON)})

	if !strings.Contains(output, "Error:") {
		t.Error("expected error for cycle detection")
	}

	if !strings.Contains(output, peb.ErrCycle.Error()) {
		t.Errorf("expected cycle error, got: %s", output)
	}
}

func TestUpdateCommandMissingID(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

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

	output := runCommand([]string{"update"})

	if !strings.Contains(output, "Error:") {
		t.Error("expected error for missing peb ID")
	}

	if !strings.Contains(output, "peb ID is required") {
		t.Errorf("expected 'peb ID is required' error, got: %s", output)
	}
}

func TestUpdateCommandPebNotFound(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

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

	update := UpdateInput{
		Title: stringPtr("New title"),
	}

	inputJSON, _ := json.Marshal(update)

	output := runCommand([]string{"update", "peb-nonexistent", string(inputJSON)})

	if !strings.Contains(output, "Error:") {
		t.Error("expected error for non-existent peb ID")
	}

	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found' error, got: %s", output)
	}
}

func TestUpdateCommandFromStdin(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	newStatus := "fixed"
	update := UpdateInput{
		Status: &newStatus,
	}

	inputJSON, _ := json.Marshal(update)
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write(inputJSON)
		w.Close()
	}()
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if updatedPeb.Status != peb.StatusFixed {
		t.Errorf("expected status fixed, got %s", updatedPeb.Status)
	}
}

func TestUpdateCommandMultipleFields(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStoreForUpdate(t)
	defer cleanup()

	id, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}

	p := peb.New(id, "Test task", peb.TypeTask, peb.StatusNew, "Initial content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	newTitle := "Updated title"
	newContent := "Updated content"
	newStatus := "in-progress"
	update := UpdateInput{
		Title:   &newTitle,
		Content: &newContent,
		Status:  &newStatus,
	}

	inputJSON, _ := json.Marshal(update)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(pebblesDir)

	output := runCommand([]string{"update", id, string(inputJSON)})

	if strings.Contains(output, "Error:") {
		t.Fatalf("unexpected error: %s", output)
	}

	s.Load()
	updatedPeb, ok := s.Get(id)
	if !ok {
		t.Fatalf("peb not found after update")
	}

	if updatedPeb.Title != newTitle {
		t.Errorf("expected title %s, got %s", newTitle, updatedPeb.Title)
	}
	if updatedPeb.Content != newContent {
		t.Errorf("expected content %s, got %s", newContent, updatedPeb.Content)
	}
	if updatedPeb.Status != peb.StatusInProgress {
		t.Errorf("expected status in-progress, got %s", updatedPeb.Status)
	}

	if !strings.Contains(output, "Updated title of "+id+".") {
		t.Errorf("expected output to contain title update message, got: %s", output)
	}
	if !strings.Contains(output, "Updated content of "+id+".") {
		t.Errorf("expected output to contain content update message, got: %s", output)
	}
	if !strings.Contains(output, "Updated status of "+id+".") {
		t.Errorf("expected output to contain status update message, got: %s", output)
	}
}

func stringPtr(s string) *string {
	return &s
}

func runCommand(args []string) string {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	app := &cli.App{
		Name:  "peb",
		Usage: "Task tracking CLI tool",
		Commands: []*cli.Command{
			InitCommand(),
			NewCommand(),
			ReadCommand(),
			UpdateCommand(),
		},
	}

	os.Args = append([]string{"peb"}, args...)
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(w, "Error:", err)
	}

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}
