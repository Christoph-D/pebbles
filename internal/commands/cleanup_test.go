package commands

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

func TestCleanupCommand(t *testing.T) {
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
	p1 := peb.New(id1, "Open peb", peb.TypeTask, peb.StatusNew, "Content 1")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p2 := peb.New(id2, "Another open peb", peb.TypeBug, peb.StatusInProgress, "Content 2")
	if err := s.Save(p2); err != nil {
		t.Fatal(err)
	}

	id3, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p3 := peb.New(id3, "Fixed peb", peb.TypeTask, peb.StatusFixed, "Content 3")
	if err := s.Save(p3); err != nil {
		t.Fatal(err)
	}

	id4, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p4 := peb.New(id4, "Wont-fix peb", peb.TypeFeature, peb.StatusWontFix, "Content 4")
	if err := s.Save(p4); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{CleanupCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "cleanup"})
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

	if !strings.Contains(lines[0], "Deleted 2") {
		t.Errorf("expected output to contain 'Deleted 2', got: %s", lines[0])
	}

	sReload := store.New(pebblesDir)
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(id1); !ok {
		t.Errorf("expected open peb %s to still exist", id1)
	}
	if _, ok := sReload.Get(id2); !ok {
		t.Errorf("expected open peb %s to still exist", id2)
	}
	if _, ok := sReload.Get(id3); ok {
		t.Errorf("expected closed peb %s to be deleted", id3)
	}
	if _, ok := sReload.Get(id4); ok {
		t.Errorf("expected closed peb %s to be deleted", id4)
	}
}

func TestCleanupCommandNoClosedPebs(t *testing.T) {
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
	p1 := peb.New(id1, "Open peb", peb.TypeTask, peb.StatusNew, "Content 1")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{CleanupCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "cleanup"})
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

	if !strings.Contains(lines[0], "No closed pebs") {
		t.Errorf("expected output to contain 'No closed pebs', got: %s", lines[0])
	}

	sReload := store.New(pebblesDir)
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if _, ok := sReload.Get(id1); !ok {
		t.Errorf("expected open peb %s to still exist", id1)
	}
}

func TestCleanupCommandAllClosed(t *testing.T) {
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
	p1 := peb.New(id1, "Fixed peb 1", peb.TypeTask, peb.StatusFixed, "Content 1")
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p2 := peb.New(id2, "Wont-fix peb", peb.TypeFeature, peb.StatusWontFix, "Content 2")
	if err := s.Save(p2); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{CleanupCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "cleanup"})
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

	if !strings.Contains(lines[0], "Deleted 2") {
		t.Errorf("expected output to contain 'Deleted 2', got: %s", lines[0])
	}

	sReload := store.New(pebblesDir)
	if err := sReload.Load(); err != nil {
		t.Fatal(err)
	}

	if len(sReload.All()) != 0 {
		t.Error("expected no pebs after cleanup")
	}
}
