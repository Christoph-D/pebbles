package commands

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/urfave/cli/v2"
)

//go:embed data/pebbles.ts
var pebblesPlugin string

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
	opencodeDir := ".opencode"
	pluginDir := filepath.Join(opencodeDir, "plugin")

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create .opencode/plugin/ directory: %w", err)
	}

	pluginPath := filepath.Join(pluginDir, "pebbles.ts")
	if err := os.WriteFile(pluginPath, []byte(pebblesPlugin), 0644); err != nil {
		return fmt.Errorf("failed to write plugin file: %w", err)
	}

	return nil
}
