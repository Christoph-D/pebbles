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
		Action: func(c *cli.Context) error {
			dir := ".pebbles"
			if _, err := os.Stat(dir); err == nil {
				return fmt.Errorf(".pebbles/ already exists in current directory.")
			}

			if err := os.Mkdir(dir, 0755); err != nil {
				return fmt.Errorf("failed to create .pebbles/ directory: %w", err)
			}

			configPath := ".pebbles/config.toml"
			if err := os.WriteFile(configPath, []byte(config.DefaultConfigContent()), 0644); err != nil {
				return fmt.Errorf("failed to create config.toml: %w", err)
			}

			fmt.Println("Initialized pebbles in .pebbles/")
			return nil
		},
	}
}
