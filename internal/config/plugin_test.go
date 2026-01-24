package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginPath(t *testing.T) {
	cfg := &Config{projectDir: "/tmp/test"}
	expected := filepath.Join("/tmp/test", ".opencode", "plugin")
	if got := pluginPath(cfg); got != expected {
		t.Errorf("pluginPath() = %v, want %v", got, expected)
	}
}

func TestReadInstalledPluginVersion(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		cfg := &Config{projectDir: tmpDir}
		pluginDir := pluginPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Version 20240123T123456Z-abc1234\n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		version, err := readInstalledPluginVersion(cfg)
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

		cfg := &Config{projectDir: tmpDir}
		pluginDir := pluginPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Version   20240123T123456Z-abc1234   \n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		version, err := readInstalledPluginVersion(cfg)
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

		cfg := &Config{projectDir: tmpDir}
		_, err := readInstalledPluginVersion(cfg)
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

		cfg := &Config{projectDir: tmpDir}
		pluginDir := pluginPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		if err := os.WriteFile(pluginFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := readInstalledPluginVersion(cfg)
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

		cfg := &Config{projectDir: tmpDir}
		pluginDir := pluginPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Some other comment\n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := readInstalledPluginVersion(cfg)
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

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		err = InstallOpencodePlugin(cfg)
		if err != nil {
			t.Fatalf("InstallOpencodePlugin() error = %v", err)
		}

		pluginDir := pluginPath(cfg)
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

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		err = InstallOpencodePlugin(cfg)
		if err != nil {
			t.Fatalf("InstallOpencodePlugin() error = %v", err)
		}

		pluginFile := filepath.Join(pluginPath(cfg), pluginFilename)
		content, err := os.ReadFile(pluginFile)
		if err != nil {
			t.Fatalf("failed to read plugin file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "task-xxxxxx") {
			t.Error("plugin file does not contain task-xxxxxx pattern (id_length=6)")
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

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		pluginFile := filepath.Join(pluginPath(cfg), pluginFilename)
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

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		pluginDir := pluginPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		oldContent := "// Version 20200101T000000Z-aaaaaaa\n// Old plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(oldContent), 0644); err != nil {
			t.Fatal(err)
		}

		err = MaybeUpdatePlugin()
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

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		pluginDir := pluginPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		newContent := "// Version 20990101T000000Z-zzzzzzz\n// New plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(newContent), 0644); err != nil {
			t.Fatal(err)
		}

		err = MaybeUpdatePlugin()
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
