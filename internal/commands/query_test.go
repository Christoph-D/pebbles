package commands

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/urfave/cli/v2"
)

func TestQueryCommand(t *testing.T) {
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
	p3 := peb.New(id3, "Third peb", peb.TypeTask, peb.StatusFixed, "Content 3")
	if err := s.Save(p3); err != nil {
		t.Fatal(err)
	}

	id4, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p4 := peb.New(id4, "Fourth peb", peb.TypeFeature, peb.StatusWontFix, "Content 4")
	if err := s.Save(p4); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    []string
		flags   map[string]string
		wantIDs []string
	}{
		{
			name:    "list all pebs",
			args:    []string{},
			wantIDs: []string{id1, id2, id3, id4},
		},
		{
			name:    "filter by status",
			args:    []string{"status:new"},
			wantIDs: []string{id1},
		},
		{
			name:    "filter by type",
			args:    []string{"type:bug"},
			wantIDs: []string{id2},
		},
		{
			name:    "multiple filters",
			args:    []string{"status:new", "type:task"},
			wantIDs: []string{id1},
		},
		{
			name:    "no results",
			args:    []string{"status:invalid"},
			wantIDs: []string{},
		},
		{
			name:    "filter by open status",
			args:    []string{"status:open"},
			wantIDs: []string{id1, id2},
		},
		{
			name:    "filter by closed status",
			args:    []string{"status:closed"},
			wantIDs: []string{id3, id4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			app := &cli.App{
				Commands: []*cli.Command{QueryCommand()},
			}

			flagArgs := []string{"query"}
			for k, v := range tt.flags {
				flagArgs = append(flagArgs, "--"+k+"="+v)
			}
			flagArgs = append(flagArgs, tt.args...)

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			err = app.Run(append([]string{"peb"}, flagArgs...))
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

			if len(lines) != len(tt.wantIDs) {
				t.Errorf("expected %d results, got %d", len(tt.wantIDs), len(lines))
			}

			for _, line := range lines {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(line), &result); err != nil {
					t.Fatalf("failed to parse JSON: %v", err)
				}

				id, ok := result["id"].(string)
				if !ok {
					t.Error("missing or invalid id field")
					continue
				}

				found := false
				for _, wantID := range tt.wantIDs {
					if id == wantID {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("unexpected ID %s in results", id)
				}
			}
		})
	}
}

func TestQueryCommandBlockedBy(t *testing.T) {
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
	blockingPeb := peb.New(blockingID, "Blocking peb", peb.TypeTask, peb.StatusNew, "Blocks another")
	if err := s.Save(blockingPeb); err != nil {
		t.Fatal(err)
	}

	id1, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p1 := peb.New(id1, "Dependent peb 1", peb.TypeTask, peb.StatusNew, "Depends on blocking")
	p1.BlockedBy = []string{blockingID}
	if err := s.Save(p1); err != nil {
		t.Fatal(err)
	}

	id2, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p2 := peb.New(id2, "Dependent peb 2", peb.TypeTask, peb.StatusNew, "Also depends")
	p2.BlockedBy = []string{blockingID, "peb-other"}
	if err := s.Save(p2); err != nil {
		t.Fatal(err)
	}

	id3, err := s.GenerateUniqueID("peb", 4)
	if err != nil {
		t.Fatal(err)
	}
	p3 := peb.New(id3, "Independent peb", peb.TypeTask, peb.StatusNew, "Not blocked")
	if err := s.Save(p3); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{QueryCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "query", "blocked-by:" + blockingID})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var results []map[string]interface{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var result map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		results = append(results, result)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	ids := make([]string, 0, 2)
	for _, result := range results {
		id, ok := result["id"].(string)
		if !ok {
			t.Error("missing or invalid id field")
			continue
		}
		ids = append(ids, id)
	}

	for _, wantID := range []string{id1, id2} {
		found := false
		for _, id := range ids {
			if id == wantID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ID %s in results", wantID)
		}
	}
}

func TestQueryCommandFields(t *testing.T) {
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
	p := peb.New(id, "Test peb", peb.TypeBug, peb.StatusInProgress, "Test content")
	p.BlockedBy = []string{"peb-blocker"}
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		fields string
		want   []string
	}{
		{
			name:   "default fields",
			fields: "",
			want:   []string{"id", "type", "status", "title"},
		},
		{
			name:   "custom fields",
			fields: "id,title,blocked-by",
			want:   []string{"id", "title", "blocked-by"},
		},
		{
			name:   "all fields",
			fields: "id,type,status,title,created,changed,blocked-by",
			want:   []string{"id", "type", "status", "title", "created", "changed", "blocked-by"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.App{
				Commands: []*cli.Command{QueryCommand()},
			}

			args := []string{"peb", "query"}
			if tt.fields != "" {
				args = append(args, "--fields="+tt.fields)
			}

			r, w, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = w

			err := app.Run(args)
			if err != nil {
				w.Close()
				os.Stdout = oldStdout
				t.Fatalf("command failed: %v", err)
			}

			w.Close()
			os.Stdout = oldStdout

			var result map[string]interface{}
			decoder := json.NewDecoder(r)
			if err := decoder.Decode(&result); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			for _, wantField := range tt.want {
				if _, ok := result[wantField]; !ok {
					t.Errorf("missing expected field %s", wantField)
				}
			}

			for field := range result {
				found := false
				for _, wantField := range tt.want {
					if field == wantField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("unexpected field %s in output", field)
				}
			}
		})
	}
}

func TestQueryCommandInvalidFilter(t *testing.T) {
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
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{QueryCommand()},
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid format",
			args: []string{"invalidfilter"},
		},
		{
			name: "unknown filter key",
			args: []string{"unknown:value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.Run([]string{"peb", "query"})
			if tt.args[0] != "" {
				err = app.Run([]string{"peb", "query", tt.args[0]})
			}
			if err == nil {
				t.Error("expected error for invalid filter")
			}
		})
	}
}

func TestQueryCommandInvalidField(t *testing.T) {
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
	p := peb.New(id, "Test peb", peb.TypeTask, peb.StatusNew, "Content")
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{QueryCommand()},
	}

	err = app.Run([]string{"peb", "query", "--fields=invalid"})
	if err == nil {
		t.Error("expected error for invalid field")
	}
}
