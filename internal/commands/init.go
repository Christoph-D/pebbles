package commands

import (
	"fmt"
	"os"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/urfave/cli/v2"
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
				if err := config.InstallOpencodePlugin(); err != nil {
					return fmt.Errorf("failed to install opencode plugin: %w", err)
				}
				fmt.Println("Installed opencode MCP plugin")
			}

			return nil
		},
	}
}
