package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const pluginFilename = "pebbles.ts"

//go:embed data/pebbles.ts
var pebblesPlugin string

//go:generate sh -c "printf '%s' $(git log -1 --format='%ct-%h' data/pebbles.ts) > data/pebbles.ts.version"
//go:embed data/pebbles.ts.version
var pebblesPluginVersionRaw string

func pebblesPluginVersion() string {
	parts := strings.Split(pebblesPluginVersionRaw, "-")
	commitEpoch, err := strconv.Atoi(parts[0])
	if len(parts) != 2 || err != nil {
		return pebblesPluginVersionRaw + "(unknown)"
	}
	timestamp := time.Unix(int64(commitEpoch), 0).UTC().Format("20060102T150405Z")
	return timestamp + "-" + parts[1]
}

func pluginPath(cfg *Config) string {
	opencodeDir := ".opencode"
	return filepath.Join(cfg.projectDir, opencodeDir, "plugin")
}

func readInstalledPluginVersion(cfg *Config) (string, error) {
	pluginFile := filepath.Join(pluginPath(cfg), pluginFilename)
	content, err := os.ReadFile(pluginFile)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("empty plugin file")
	}

	line := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(line, "// Version ") {
		return "", fmt.Errorf("plugin file missing version comment")
	}

	version := strings.TrimPrefix(line, "// Version ")
	return strings.TrimSpace(version), nil
}

func MaybeUpdatePlugin() error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	installedVersion, err := readInstalledPluginVersion(cfg)
	if err != nil {
		return nil
	}

	currentVersion := pebblesPluginVersion()

	if currentVersion > installedVersion {
		return InstallOpencodePlugin(cfg)
	}

	return nil
}

func InstallOpencodePlugin(cfg *Config) error {
	tmpl, err := template.New("pebblesPlugin").Parse(pebblesPlugin)
	if err != nil {
		return err
	}

	data := struct {
		PebbleIDSuffix   string
		PebbleIDPattern  string
		PebbleIDPattern2 string
		PebbleIDPattern3 string
		Version          string
	}{
		PebbleIDSuffix:   strings.Repeat("x", cfg.IDLength),
		PebbleIDPattern:  cfg.Prefix + "-" + strings.Repeat("x", cfg.IDLength),
		PebbleIDPattern2: cfg.Prefix + "-" + strings.Repeat("y", cfg.IDLength),
		PebbleIDPattern3: cfg.Prefix + "-" + strings.Repeat("z", cfg.IDLength),
		Version:          pebblesPluginVersion(),
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	pluginDir := pluginPath(cfg)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create .opencode/plugin/ directory: %w", err)
	}

	pluginFile := filepath.Join(pluginDir, pluginFilename)
	if err := os.WriteFile(pluginFile, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write plugin file: %w", err)
	}

	return nil
}
