package peb

import (
	"strings"
	"testing"
)

func TestFilenameLengthLimit(t *testing.T) {
	p := &Peb{
		ID:    "peb-abcd",
		Title: "this is a very very very very very very very very very very very very very very very very very very very very very very very very very very long title that should be truncated",
		Type:  TypeBug,
	}

	filename := Filename(p)
	if len(filename) > 100 {
		t.Errorf("filename length %d exceeds 100 character limit: %s", len(filename), filename)
	}
}

func TestFilenameNormalTitle(t *testing.T) {
	p := &Peb{
		ID:    "peb-abcd",
		Title: "Fix login bug",
		Type:  TypeBug,
	}

	filename := Filename(p)
	expected := "peb-peb-abcd--fix-login-bug.md"
	if filename != expected {
		t.Errorf("expected %s, got %s", expected, filename)
	}
}

func TestFilenameEmptyTitle(t *testing.T) {
	p := &Peb{
		ID:    "peb-abcd",
		Title: "",
		Type:  TypeBug,
	}

	filename := Filename(p)
	if !strings.HasSuffix(filename, "--untitled.md") {
		t.Errorf("expected filename to end with --untitled.md, got %s", filename)
	}
}
