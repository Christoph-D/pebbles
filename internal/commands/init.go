package commands

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"go.yozora.eu/pebbles/internal/config"
)

func InitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new pebbles project",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "opencode",
				Usage: "Install or update opencode MCP plugin (overwrites existing plugin file)",
			},
			&cli.BoolFlag{
				Name:  "pi",
				Usage: "Install or update pi agent extension (overwrites existing extension file)",
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

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if err := config.MaybeUpdatePlugin(cfg); err != nil {
				return fmt.Errorf("failed to update plugin: %w", err)
			}

			if c.Bool("opencode") {
				if err := config.InstallOpencodePlugin(cfg); err != nil {
					return fmt.Errorf("failed to install opencode plugin: %w", err)
				}
				fmt.Println("Installed opencode MCP plugin")
			}

			if c.Bool("pi") {
				if err := config.InstallPiExtension(cfg); err != nil {
					return fmt.Errorf("failed to install pi extension: %w", err)
				}
				fmt.Println("Installed pi agent extension")
			}

			return nil
		},
	}
}
