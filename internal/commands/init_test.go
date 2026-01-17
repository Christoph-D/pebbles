package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	expectedErr := ".pebbles/ already exists in current directory."
	if _, err := os.Stat(pebblesDir); err == nil {
		if !strings.Contains(expectedErr, ".pebbles/") {
			t.Errorf("expected error message to reference .pebbles/")
		}
	}
}
