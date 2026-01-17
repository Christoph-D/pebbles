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
	expected := "peb-abcd--fix-login-bug.md"
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

func TestParseID(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		prefix   string
		want     string
		wantErr  bool
	}{
		{
			name:     "normal filename",
			filename: "peb-abcd--fix-login-bug.md",
			prefix:   "peb",
			want:     "peb-abcd",
			wantErr:  false,
		},
		{
			name:     "custom prefix",
			filename: "task-xyz--do-something.md",
			prefix:   "task",
			want:     "task-xyz",
			wantErr:  false,
		},
		{
			name:     "wrong prefix",
			filename: "peb-abcd--fix-login-bug.md",
			prefix:   "task",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseID(tt.filename, tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseID() = %v, want %v", got, tt.want)
			}
		})
	}
}
