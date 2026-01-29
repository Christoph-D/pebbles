package commands

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestConfigCommand(t *testing.T) {
	pebblesDir, _, cleanup := setupTestStore(t)
	defer cleanup()
	t.Chdir(pebblesDir)

	app := &cli.App{
		Commands: []*cli.Command{ConfigCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := app.Run([]string{"peb", "config"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var output string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if output != "" {
			output += "\n"
		}
		output += scanner.Text()
	}

	if output == "" {
		t.Fatal("expected output, got empty")
	}

	type ConfigOutput struct {
		Prefix   string `json:"prefix"`
		IDLength int    `json:"id_length"`
	}

	var config ConfigOutput
	if err := json.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if config.Prefix != "peb" {
		t.Errorf("expected prefix 'peb', got '%s'", config.Prefix)
	}

	if config.IDLength != 4 {
		t.Errorf("expected ID length 4, got %d", config.IDLength)
	}
}

func TestConfigCommandCustomConfig(t *testing.T) {
	pebblesDir, _, cleanup := setupTestStore(t)
	defer cleanup()
	t.Chdir(pebblesDir)

	configPath := pebblesDir + "/config.toml"
	configContent := `prefix = "task"
id_length = 6
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{ConfigCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := app.Run([]string{"peb", "config"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var output string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if output != "" {
			output += "\n"
		}
		output += scanner.Text()
	}

	type ConfigOutput struct {
		Prefix   string `json:"prefix"`
		IDLength int    `json:"id_length"`
	}

	var config ConfigOutput
	if err := json.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if config.Prefix != "task" {
		t.Errorf("expected prefix 'task', got '%s'", config.Prefix)
	}

	if config.IDLength != 6 {
		t.Errorf("expected ID length 6, got %d", config.IDLength)
	}
}

func TestConfigCommandJSONFormatting(t *testing.T) {
	pebblesDir, _, cleanup := setupTestStore(t)
	defer cleanup()
	t.Chdir(pebblesDir)

	app := &cli.App{
		Commands: []*cli.Command{ConfigCommand()},
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := app.Run([]string{"peb", "config"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var output string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if output != "" {
			output += "\n"
		}
		output += scanner.Text()
	}

	if !strings.Contains(output, "  ") {
		t.Error("expected JSON to be indented, got flat output")
	}

	type ConfigOutput struct {
		Prefix   string `json:"prefix"`
		IDLength int    `json:"id_length"`
	}

	var config ConfigOutput
	if err := json.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if config.Prefix != "peb" {
		t.Errorf("expected prefix 'peb', got '%s'", config.Prefix)
	}
}
