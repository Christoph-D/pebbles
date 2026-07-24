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
var pebblesOpencodePlugin string

//go:generate sh -c "printf '%s' $(git log -1 --format='%ct-%h' data/pebbles.ts) > data/pebbles.ts.version"
//go:embed data/pebbles.ts.version
var pebblesOpencodePluginVersion string

//go:embed data/pebbles-pi.ts
var pebblesPiPlugin string

//go:generate sh -c "printf '%s' $(git log -1 --format='%ct-%h' data/pebbles-pi.ts) > data/pebbles-pi.ts.version"
//go:embed data/pebbles-pi.ts.version
var pebblesPiPluginVersion string

// pluginSpec describes an auto-generated agent integration file.
type pluginSpec struct {
	name       string // human-readable identifier, e.g. "opencode" or "pi"
	dir        string // directory relative to the project root
	template   string // embedded file content (a Go text/template)
	versionRaw string // embedded version file content ("<epoch>-<short-hash>")
}

// version renders the spec's version string from its raw version file content.
func (s pluginSpec) version() string {
	parts := strings.Split(s.versionRaw, "-")
	commitEpoch, err := strconv.Atoi(parts[0])
	if len(parts) != 2 || err != nil {
		return s.versionRaw + "(unknown)"
	}
	timestamp := time.Unix(int64(commitEpoch), 0).UTC().Format("20060102T150405Z")
	return timestamp + "-" + parts[1]
}

func (s pluginSpec) dirPath(cfg *Config) string {
	return filepath.Join(cfg.projectDir, s.dir)
}

func (s pluginSpec) filePath(cfg *Config) string {
	return filepath.Join(s.dirPath(cfg), pluginFilename)
}

// readInstalledVersion reads the version comment from the first line of the
// installed plugin file.
func (s pluginSpec) readInstalledVersion(cfg *Config) (string, error) {
	content, err := os.ReadFile(s.filePath(cfg))
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

	return strings.TrimSpace(strings.TrimPrefix(line, "// Version ")), nil
}

// install renders the spec's template and writes it to the project, creating
// the target directory if needed.
func (s pluginSpec) install(cfg *Config) error {
	tmpl, err := template.New(s.name + "Plugin").Parse(s.template)
	if err != nil {
		return err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, struct {
		Version string
	}{
		Version: s.version(),
	}); err != nil {
		return err
	}

	if err := os.MkdirAll(s.dirPath(cfg), 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", s.dir, err)
	}

	if err := os.WriteFile(s.filePath(cfg), []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write %s plugin file: %w", s.name, err)
	}

	return nil
}

// maybeUpdate rewrites the plugin file if it already exists and the embedded
// template is newer than the installed version. It does nothing when the
// plugin is not installed.
func (s pluginSpec) maybeUpdate(cfg *Config) error {
	installedVersion, err := s.readInstalledVersion(cfg)
	if err != nil {
		return nil
	}

	if s.version() > installedVersion {
		return s.install(cfg)
	}

	return nil
}

var opencodePluginSpec = pluginSpec{
	name:       "opencode",
	dir:        ".opencode/plugin",
	template:   pebblesOpencodePlugin,
	versionRaw: pebblesOpencodePluginVersion,
}

var piPluginSpec = pluginSpec{
	name:       "pi",
	dir:        ".pi/extensions",
	template:   pebblesPiPlugin,
	versionRaw: pebblesPiPluginVersion,
}

// allPluginSpecs is the full set of agent integrations pebbles manages.
var allPluginSpecs = []pluginSpec{opencodePluginSpec, piPluginSpec}

// MaybeUpdatePlugin updates any already-installed agent plugin whose embedded
// template is newer than the installed version. Plugins that are not present
// are left alone.
func MaybeUpdatePlugin(cfg *Config) error {
	for _, spec := range allPluginSpecs {
		if err := spec.maybeUpdate(cfg); err != nil {
			return err
		}
	}
	return nil
}

// InstallOpencodePlugin writes the opencode MCP plugin to .opencode/plugin/.
func InstallOpencodePlugin(cfg *Config) error {
	return opencodePluginSpec.install(cfg)
}

// InstallPiExtension writes the pi agent extension to .pi/extensions/.
func InstallPiExtension(cfg *Config) error {
	return piPluginSpec.install(cfg)
}
