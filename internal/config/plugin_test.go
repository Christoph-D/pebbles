package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginSpecDirPath(t *testing.T) {
	cfg := &Config{projectDir: "/tmp/test"}
	if got := opencodePluginSpec.dirPath(cfg); got != filepath.Join("/tmp/test", ".opencode", "plugin") {
		t.Errorf("opencodePluginSpec.dirPath() = %v, want %v", got, filepath.Join("/tmp/test", ".opencode", "plugin"))
	}
	if got := piPluginSpec.dirPath(cfg); got != filepath.Join("/tmp/test", ".pi", "extensions") {
		t.Errorf("piPluginSpec.dirPath() = %v, want %v", got, filepath.Join("/tmp/test", ".pi", "extensions"))
	}
}

func TestReadInstalledPluginVersion(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		cfg := &Config{projectDir: tmpDir}
		pluginDir := opencodePluginSpec.dirPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Version 20240123T123456Z-abc1234\n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		version, err := opencodePluginSpec.readInstalledVersion(cfg)
		if err != nil {
			t.Fatalf("readInstalledVersion() error = %v", err)
		}

		expected := "20240123T123456Z-abc1234"
		if version != expected {
			t.Errorf("readInstalledVersion() = %v, want %v", version, expected)
		}
	})

	t.Run("version with leading/trailing spaces", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		cfg := &Config{projectDir: tmpDir}
		pluginDir := opencodePluginSpec.dirPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Version   20240123T123456Z-abc1234   \n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		version, err := opencodePluginSpec.readInstalledVersion(cfg)
		if err != nil {
			t.Fatalf("readInstalledVersion() error = %v", err)
		}

		expected := "20240123T123456Z-abc1234"
		if version != expected {
			t.Errorf("readInstalledVersion() = %v, want %v", version, expected)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		cfg := &Config{projectDir: tmpDir}
		_, err := opencodePluginSpec.readInstalledVersion(cfg)
		if err == nil {
			t.Fatal("readInstalledVersion() expected error for missing file")
		}
		if !os.IsNotExist(err) {
			t.Errorf("readInstalledVersion() error = %v, want IsNotExist error", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		cfg := &Config{projectDir: tmpDir}
		pluginDir := opencodePluginSpec.dirPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		if err := os.WriteFile(pluginFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := opencodePluginSpec.readInstalledVersion(cfg)
		if err == nil {
			t.Fatal("readInstalledVersion() expected error for empty file")
		}
		if !strings.Contains(err.Error(), "missing version comment") {
			t.Errorf("readInstalledVersion() error = %v, want missing version comment error", err)
		}
	})

	t.Run("missing version comment", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		cfg := &Config{projectDir: tmpDir}
		pluginDir := opencodePluginSpec.dirPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		content := "// Some other comment\n// Plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := opencodePluginSpec.readInstalledVersion(cfg)
		if err == nil {
			t.Fatal("readInstalledVersion() expected error for missing version comment")
		}
		if !strings.Contains(err.Error(), "missing version comment") {
			t.Errorf("readInstalledVersion() error = %v, want missing version comment error", err)
		}
	})
}

func TestInstallOpencodePlugin(t *testing.T) {
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

	pluginDir := opencodePluginSpec.dirPath(cfg)
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
}

func TestInstallPiExtension(t *testing.T) {
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

	if err := InstallPiExtension(cfg); err != nil {
		t.Fatalf("InstallPiExtension() error = %v", err)
	}

	pluginDir := piPluginSpec.dirPath(cfg)
	pluginFile := filepath.Join(pluginDir, pluginFilename)

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Fatal(".pi/extensions directory was not created")
	}

	content, err := os.ReadFile(pluginFile)
	if err != nil {
		t.Fatalf("failed to read extension file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "registerTool") {
		t.Error("pi extension file does not contain registerTool call")
	}
	if !strings.Contains(contentStr, "peb_query") {
		t.Error("pi extension file does not contain peb_query tool")
	}
	if !strings.Contains(contentStr, "// Version ") {
		t.Error("pi extension file does not contain version comment")
	}
	if !strings.HasPrefix(contentStr, "// Version ") {
		t.Error("pi extension file should start with version comment")
	}

	// The pi extension installs into .pi/extensions, never .opencode/plugin.
	opencodeFile := filepath.Join(opencodePluginSpec.dirPath(cfg), pluginFilename)
	if _, err := os.Stat(opencodeFile); !os.IsNotExist(err) {
		t.Error("installing pi extension should not create an opencode plugin file")
	}
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

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		err = MaybeUpdatePlugin(cfg)
		if err != nil {
			t.Fatalf("MaybeUpdatePlugin() error = %v", err)
		}
		for _, spec := range allPluginSpecs {
			pluginFile := filepath.Join(spec.dirPath(cfg), pluginFilename)
			if _, err := os.Stat(pluginFile); !os.IsNotExist(err) {
				t.Errorf("plugin file should not be installed when it doesn't exist: %s", spec.name)
			}
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

		pluginDir := opencodePluginSpec.dirPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		oldContent := "// Version 20200101T000000Z-aaaaaaa\n// Old plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(oldContent), 0644); err != nil {
			t.Fatal(err)
		}

		err = MaybeUpdatePlugin(cfg)
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

		pluginDir := opencodePluginSpec.dirPath(cfg)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		pluginFile := filepath.Join(pluginDir, pluginFilename)
		newContent := "// Version 20990101T000000Z-zzzzzzz\n// New plugin content\n"
		if err := os.WriteFile(pluginFile, []byte(newContent), 0644); err != nil {
			t.Fatal(err)
		}

		err = MaybeUpdatePlugin(cfg)
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

	t.Run("updates any installed spec with older version", func(t *testing.T) {
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

		// A synthetic spec with a known "current" version far in the future.
		// This tests the shared maybeUpdate comparison logic independent of the
		// real specs' git-derived versions.
		spec := pluginSpec{
			name:       "test-agent",
			dir:        ".test-agent/ext",
			template:   "// Version {{.Version}}\n// test-agent body\n",
			versionRaw: "9999999999-zzzzzzz",
		}

		if err := os.MkdirAll(spec.dirPath(cfg), 0755); err != nil {
			t.Fatal(err)
		}
		pluginFile := spec.filePath(cfg)
		oldContent := "// Version 1000000000-aaaaaaa\n// old test-agent content\n"
		if err := os.WriteFile(pluginFile, []byte(oldContent), 0644); err != nil {
			t.Fatal(err)
		}

		if err := spec.maybeUpdate(cfg); err != nil {
			t.Fatalf("maybeUpdate() error = %v", err)
		}

		content, err := os.ReadFile(pluginFile)
		if err != nil {
			t.Fatalf("failed to read plugin file: %v", err)
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "old test-agent content") {
			t.Error("plugin file was not updated")
		}
		if !strings.Contains(contentStr, "test-agent body") {
			t.Error("plugin file does not contain rendered template body")
		}
		if !strings.Contains(contentStr, "// Version ") || !strings.Contains(contentStr, "zzzzzzz") {
			t.Error("plugin file does not contain rendered version")
		}
	})
}
