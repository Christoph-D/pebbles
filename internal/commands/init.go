package commands

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/urfave/cli/v2"
)

//go:embed data/pebbles.ts
var pebblesPlugin string

//go:generate sh -c "printf '%s' $(git log -1 --format='%ct-%h' data/pebbles.ts) > data/pebbles.ts.version"
//go:embed data/pebbles.ts.version
var pebblesPluginVersionRaw string

func pebblesPluginVersion() string {
	parts := strings.Split(pebblesPluginVersionRaw, "-")
	commitEpoch, err := strconv.Atoi(parts[0])
	if len(parts) != 2 || err != nil {
		return pebblesPluginVersionRaw + "-unknown"
	}
	timestamp := time.Unix(int64(commitEpoch), 0).UTC().Format("20060102T150405Z")
	return timestamp + "-" + parts[1]
}

func InitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new pebbles project",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "opencode",
				Usage: "Install or update opencode MCP plugin (overwrites existing plugin file)",
			},
		},
		Action: func(c *cli.Context) error {
			dir := ".pebbles"
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create .pebbles/ directory: %w", err)
			}

			configPath := ".pebbles/config.toml"
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				if err := os.WriteFile(configPath, []byte(config.DefaultConfigContent()), 0644); err != nil {
					return fmt.Errorf("failed to create config.toml: %w", err)
				}
				fmt.Println("Initialized pebbles in .pebbles/")
			}

			if c.Bool("opencode") {
				if err := installOpencodePlugin(); err != nil {
					return fmt.Errorf("failed to install opencode plugin: %w", err)
				}
				fmt.Println("Installed opencode MCP plugin")
			}

			return nil
		},
	}
}

func installOpencodePlugin() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

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

	opencodeDir := ".opencode"
	pluginDir := filepath.Join(opencodeDir, "plugin")

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create .opencode/plugin/ directory: %w", err)
	}

	pluginPath := filepath.Join(pluginDir, "pebbles.ts")
	if err := os.WriteFile(pluginPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write plugin file: %w", err)
	}

	return nil
}
