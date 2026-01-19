package commands

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/urfave/cli/v2"
)

func TestDeleteCommand(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
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

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "delete", id})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 1 {
		t.Fatalf("expected 1 output line, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Deleted peb") {
		t.Errorf("expected output to contain 'Deleted peb', got: %s", lines[0])
	}

	sReload := s
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(id); ok {
		t.Errorf("expected peb %s to be deleted", id)
	}
}

func TestDeleteCommandMultipleIDs(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
		t.Fatal(err)
	}

	id1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p1 := peb.New(id1, "First peb", peb.TypeTask, peb.StatusNew, "Content 1")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p2 := peb.New(id2, "Second peb", peb.TypeBug, peb.StatusInProgress, "Content 2")
	if err := s.Save(p2); err != nil {
		t.Fatal(err)
	}

	id3, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p3 := peb.New(id3, "Third peb", peb.TypeFeature, peb.StatusFixed, "Content 3")
	if err := s.Save(p3); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "delete", id1, id2, id3})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 1 {
		t.Fatalf("expected 1 output line, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Deleted pebs") {
		t.Errorf("expected output to contain 'Deleted pebs', got: %s", lines[0])
	}

	sReload := s
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(id1); ok {
		t.Errorf("expected peb %s to be deleted", id1)
	}
	if _, ok := sReload.Get(id2); ok {
		t.Errorf("expected peb %s to be deleted", id2)
	}
	if _, ok := sReload.Get(id3); ok {
		t.Errorf("expected peb %s to be deleted", id3)
	}
}

func TestDeleteCommandNotFound(t *testing.T) {
	pebblesDir, _, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	err = app.Run([]string{"peb", "delete", "peb-nonexistent"})
	if err == nil {
		t.Error("expected error for non-existent peb, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got: %v", err)
	}
}

func TestDeleteCommandNoArgs(t *testing.T) {
	pebblesDir, _, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	err = app.Run([]string{"peb", "delete"})
	if err == nil {
		t.Error("expected error for missing peb ID, got nil")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected error to contain 'required', got: %v", err)
	}
}

func TestDeleteCommandWithDependentPebsNotBeingDeleted(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
		t.Fatal(err)
	}

	blockingID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	blockingPeb := peb.New(blockingID, "Blocking peb", peb.TypeTask, peb.StatusNew, "Blocks another peb")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	dependentID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	dependentPeb := peb.New(dependentID, "Dependent peb", peb.TypeTask, peb.StatusNew, "Depends on blocking peb")
	dependentPeb.BlockedBy = []string{blockingID}
	if err := s.Save(dependentPeb); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	err = app.Run([]string{"peb", "delete", blockingID})
	if err == nil {
		t.Error("expected error when deleting peb with dependent pebs not being deleted, got nil")
	}

	if !strings.Contains(err.Error(), "referenced by blocked-by") {
		t.Errorf("expected error to contain 'referenced by blocked-by', got: %v", err)
	}

	sReload := s
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(blockingID); !ok {
		t.Errorf("expected blocking peb %s to still exist", blockingID)
	}
	if _, ok := sReload.Get(dependentID); !ok {
		t.Errorf("expected dependent peb %s to still exist", dependentID)
	}
}

func TestDeleteCommandWithDependentPebsBeingDeleted(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
		t.Fatal(err)
	}

	blockingID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	blockingPeb := peb.New(blockingID, "Blocking peb", peb.TypeTask, peb.StatusNew, "Blocks another peb")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	dependentID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	dependentPeb := peb.New(dependentID, "Dependent peb", peb.TypeTask, peb.StatusNew, "Depends on blocking peb")
	dependentPeb.BlockedBy = []string{blockingID}
	if err := s.Save(dependentPeb); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "delete", blockingID, dependentID})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 1 {
		t.Fatalf("expected 1 output line, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Deleted pebs") {
		t.Errorf("expected output to contain 'Deleted pebs', got: %s", lines[0])
	}

	sReload := s
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(blockingID); ok {
		t.Errorf("expected blocking peb %s to be deleted", blockingID)
	}
	if _, ok := sReload.Get(dependentID); ok {
		t.Errorf("expected dependent peb %s to be deleted", dependentID)
	}
}

func TestDeleteCommandPartialFailure(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
		t.Fatal(err)
	}

	id1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p1 := peb.New(id1, "Existing peb", peb.TypeTask, peb.StatusNew, "Content 1")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	err = app.Run([]string{"peb", "delete", id1, "peb-nonexistent"})
	if err == nil {
		t.Error("expected error for non-existent peb, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got: %v", err)
	}

	sReload := s
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(id1); !ok {
		t.Error("expected existing peb to still exist after partial failure")
	}
}

func TestDeleteCommandWithSomeDependantsNotBeingDeleted(t *testing.T) {
	pebblesDir, s, cleanup := setupTestStore(t)
	defer cleanup()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(pebblesDir); err != nil {
		t.Fatal(err)
	}

	blockingID, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	blockingPeb := peb.New(blockingID, "Blocking peb", peb.TypeTask, peb.StatusNew, "Blocks two pebs")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	dependentID1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	dependentPeb1 := peb.New(dependentID1, "Dependent peb 1", peb.TypeTask, peb.StatusNew, "Depends on blocking peb")
	dependentPeb1.BlockedBy = []string{blockingID}
	if err := s.Save(dependentPeb1); err != nil {
		t.Fatal(err)
	}

	dependentID2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	dependentPeb2 := peb.New(dependentID2, "Dependent peb 2", peb.TypeTask, peb.StatusNew, "Also depends on blocking peb")
	dependentPeb2.BlockedBy = []string{blockingID}
	if err := s.Save(dependentPeb2); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{DeleteCommand()},
	}

	err = app.Run([]string{"peb", "delete", blockingID, dependentID1})
	if err == nil {
		t.Error("expected error when deleting peb with some dependants not being deleted, got nil")
	}

	if !strings.Contains(err.Error(), "referenced by blocked-by") {
		t.Errorf("expected error to contain 'referenced by blocked-by', got: %v", err)
	}

	if !strings.Contains(err.Error(), "not being deleted") {
		t.Errorf("expected error to contain 'not being deleted', got: %v", err)
	}

	if !strings.Contains(err.Error(), dependentID2) {
		t.Errorf("expected error to mention dependent peb not being deleted %s, got: %v", dependentID2, err)
	}

	sReload := s
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(blockingID); !ok {
		t.Errorf("expected blocking peb %s to still exist", blockingID)
	}
	if _, ok := sReload.Get(dependentID1); !ok {
		t.Errorf("expected dependent peb %s to still exist", dependentID1)
	}
	if _, ok := sReload.Get(dependentID2); !ok {
		t.Errorf("expected dependent peb %s to still exist", dependentID2)
	}
}
