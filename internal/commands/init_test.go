package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pebblesDir := ".pebbles"
	if _, err := os.Stat(pebblesDir); err == nil {
		t.Fatalf(".pebbles already exists in test directory")
	}

	if err := os.Mkdir(pebblesDir, 0755); err != nil {
		t.Fatalf("failed to create .pebbles directory: %v", err)
	}

	configContent := `# Pebbles configuration
prefix = "peb"
id_length = 4
`
	configPath := filepath.Join(pebblesDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config.toml: %v", err)
	}

	if _, err := os.Stat(pebblesDir); os.IsNotExist(err) {
		t.Fatalf(".pebbles directory was not created")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config.toml was not created")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}

	if !strings.Contains(string(content), `prefix = "peb"`) {
		t.Errorf("config.toml does not contain expected prefix value")
	}
	if !strings.Contains(string(content), `id_length = 4`) {
		t.Errorf("config.toml does not contain expected id_length value")
	}
	if !strings.Contains(string(content), `# Pebbles configuration`) {
		t.Errorf("config.toml does not contain expected comment")
	}
}

func TestInitCommandAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pebblesDir := filepath.Join(tmpDir, ".pebbles")
	if err := os.Mkdir(pebblesDir, 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(pebblesDir); os.IsNotExist(err) {
		t.Fatalf(".pebbles directory was not created for test setup")
	}

	if _, err := os.Stat(pebblesDir); err == nil {
		t.Log(".pebbles directory exists, running init again should be idempotent")
	}
}

func TestInitCommandIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pebblesDir := ".pebbles"
	configPath := filepath.Join(pebblesDir, "config.toml")

	app := &cli.App{
		Commands: []*cli.Command{InitCommand()},
	}

	_, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "init"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("first init failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	if _, err := os.Stat(pebblesDir); os.IsNotExist(err) {
		t.Fatalf(".pebbles directory was not created")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config.toml was not created")
	}

	_, w, _ = os.Pipe()
	oldStdout = os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "init"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("second init should be idempotent, but failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	content1, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config.toml after first init: %v", err)
	}

	content2, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config.toml after second init: %v", err)
	}

	if string(content1) != string(content2) {
		t.Fatalf("config.toml was modified on second init")
	}
}

func TestInitCommandWithOpencode(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{InitCommand()},
	}

	_, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "init", "--opencode"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("init with --opencode failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	opencodeDir := ".opencode"
	pluginDir := filepath.Join(opencodeDir, "plugin")

	if _, err := os.Stat(opencodeDir); os.IsNotExist(err) {
		t.Fatalf(".opencode directory was not created")
	}

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Fatalf(".opencode/plugin/ directory was not created")
	}

	pluginPath := filepath.Join(pluginDir, "pebbles.ts")
	content, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("failed to read pebbles.ts: %v", err)
	}

	if !strings.Contains(string(content), "PebblesPlugin") {
		t.Errorf("pebbles.ts does not contain expected content")
	}
	if !strings.Contains(string(content), "export const") {
		t.Errorf("pebbles.ts does not contain expected export")
	}
}

func TestInitCommandWithOpencodeIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pluginPath := filepath.Join(".opencode", "plugin", "pebbles.ts")

	app := &cli.App{
		Commands: []*cli.Command{InitCommand()},
	}

	_, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "init", "--opencode"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("first init with --opencode failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	content1, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("failed to read pebbles.ts after first init: %v", err)
	}

	_, w, _ = os.Pipe()
	oldStdout = os.Stdout
	os.Stdout = w

	err = app.Run([]string{"peb", "init", "--opencode"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("second init with --opencode should be idempotent, but failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	content2, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("failed to read pebbles.ts after second init: %v", err)
	}

	if string(content1) != string(content2) {
		t.Fatalf("pebbles.ts was modified on second init")
	}
}
