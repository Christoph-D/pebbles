package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginPath(t *testing.T) {
	expected := filepath.Join(".opencode", "plugin")
	if got := pluginPath(); got != expected {
		t.Errorf("pluginPath() = %v, want %v", got, expected)
	}
}

func TestReadInstalledPluginVersion(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pluginDir := pluginPath()
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Version 20240123T123456Z-abc1234\n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		version, err := readInstalledPluginVersion()
		if err != nil {
			t.Fatalf("readInstalledPluginVersion() error = %v", err)
		}

		expected := "20240123T123456Z-abc1234"
		if version != expected {
			t.Errorf("readInstalledPluginVersion() = %v, want %v", version, expected)
		}
	})

	t.Run("version with leading/trailing spaces", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pluginDir := pluginPath()
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Version   20240123T123456Z-abc1234   \n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		version, err := readInstalledPluginVersion()
		if err != nil {
			t.Fatalf("readInstalledPluginVersion() error = %v", err)
		}

		expected := "20240123T123456Z-abc1234"
		if version != expected {
			t.Errorf("readInstalledPluginVersion() = %v, want %v", version, expected)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		_, err := readInstalledPluginVersion()
		if err == nil {
			t.Fatal("readInstalledPluginVersion() expected error for missing file")
		}
		if !os.IsNotExist(err) {
			t.Errorf("readInstalledPluginVersion() error = %v, want IsNotExist error", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pluginDir := pluginPath()
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		if err := os.WriteFile(pluginFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := readInstalledPluginVersion()
		if err == nil {
			t.Fatal("readInstalledPluginVersion() expected error for empty file")
		}
		if !strings.Contains(err.Error(), "missing version comment") {
			t.Errorf("readInstalledPluginVersion() error = %v, want missing version comment error", err)
		}
	})

	t.Run("missing version comment", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pluginDir := pluginPath()
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Some other comment\n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := readInstalledPluginVersion()
		if err == nil {
			t.Fatal("readInstalledPluginVersion() expected error for missing version comment")
		}
		if !strings.Contains(err.Error(), "missing version comment") {
			t.Errorf("readInstalledPluginVersion() error = %v, want missing version comment error", err)
		}
	})
}

func TestInstallOpencodePlugin(t *testing.T) {
	t.Run("successful installation", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pebblesDir := ".pebbles"
		if err := os.Mkdir(pebblesDir, 0755); err != nil {
			t.Fatal(err)
		}

		configContent := `# Pebbles configuration
prefix = "peb"
id_length = 4
`
		configPath := filepath.Join(pebblesDir, "config.toml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		err := InstallOpencodePlugin()
		if err != nil {
			t.Fatalf("InstallOpencodePlugin() error = %v", err)
		}

		pluginDir := pluginPath()
		pluginFile := filepath.Join(pluginDir, pluginFilename)

		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			t.Fatal("plugin directory was not created")
		}

		content, err := os.ReadFile(pluginFile)
		if err != nil {
			t.Fatalf("failed to read plugin file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "peb_new") {
			t.Error("plugin file does not contain peb_new tool")
		}
		if !strings.Contains(contentStr, "// Version ") {
			t.Error("plugin file does not contain version comment")
		}
		if !strings.Contains(contentStr, "peb-xxxx") {
			t.Error("plugin file does not contain peb-xxxx pattern (id_length=4)")
		}
	})

	t.Run("custom id_length", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pebblesDir := ".pebbles"
		if err := os.Mkdir(pebblesDir, 0755); err != nil {
			t.Fatal(err)
		}

		configContent := `# Pebbles configuration
prefix = "task"
id_length = 6
`
		configPath := filepath.Join(pebblesDir, "config.toml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		err := InstallOpencodePlugin()
		if err != nil {
			t.Fatalf("InstallOpencodePlugin() error = %v", err)
		}

		pluginFile := filepath.Join(pluginPath(), pluginFilename)
		content, err := os.ReadFile(pluginFile)
		if err != nil {
			t.Fatalf("failed to read plugin file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "task-xxxxxx") {
			t.Error("plugin file does not contain task-xxxxxx pattern (id_length=6)")
		}
	})

	t.Run("no config directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		err := InstallOpencodePlugin()
		if err == nil {
			t.Fatal("InstallOpencodePlugin() expected error for missing .pebbles directory")
		}
	})
}

func TestMaybeUpdatePlugin(t *testing.T) {
	t.Run("no existing plugin returns nil without installing", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pebblesDir := ".pebbles"
		if err := os.Mkdir(pebblesDir, 0755); err != nil {
			t.Fatal(err)
		}

		configContent := `# Pebbles configuration
prefix = "peb"
id_length = 4
`
		configPath := filepath.Join(pebblesDir, "config.toml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		err := MaybeUpdatePlugin()
		if err != nil {
			t.Fatalf("MaybeUpdatePlugin() error = %v", err)
		}

		pluginFile := filepath.Join(pluginPath(), pluginFilename)
		if _, err := os.Stat(pluginFile); !os.IsNotExist(err) {
			t.Error("plugin file should not be installed when it doesn't exist")
		}
	})

	t.Run("existing plugin with older version gets updated", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pebblesDir := ".pebbles"
		if err := os.Mkdir(pebblesDir, 0755); err != nil {
			t.Fatal(err)
		}

		configContent := `# Pebbles configuration
prefix = "peb"
id_length = 4
`
		configPath := filepath.Join(pebblesDir, "config.toml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		pluginDir := pluginPath()
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		oldContent := "// Version 20200101T000000Z-aaaaaaa\n// Old plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(oldContent), 0644); err != nil {
			t.Fatal(err)
		}

		err := MaybeUpdatePlugin()
		if err != nil {
			t.Fatalf("MaybeUpdatePlugin() error = %v", err)
		}

		content, err := os.ReadFile(pluginFile)
		if err != nil {
			t.Fatalf("failed to read plugin file: %v", err)
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "Old plugin content") {
			t.Error("plugin file was not updated")
		}
	})

	t.Run("existing plugin with newer version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		pebblesDir := ".pebbles"
		if err := os.Mkdir(pebblesDir, 0755); err != nil {
			t.Fatal(err)
		}

		configContent := `# Pebbles configuration
prefix = "peb"
id_length = 4
`
		configPath := filepath.Join(pebblesDir, "config.toml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		pluginDir := pluginPath()
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		newContent := "// Version 20990101T000000Z-zzzzzzz\n// New plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(newContent), 0644); err != nil {
			t.Fatal(err)
		}

		err := MaybeUpdatePlugin()
		if err != nil {
			t.Fatalf("MaybeUpdatePlugin() error = %v", err)
		}

		content, err := os.ReadFile(pluginFile)
		if err != nil {
			t.Fatalf("failed to read plugin file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "New plugin content") {
			t.Error("plugin file was incorrectly updated")
		}
	})
}
