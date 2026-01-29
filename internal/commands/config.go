package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/urfave/cli/v2"
)

func ConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Display configuration as JSON",
		Description: `Display the current pebbles configuration as formatted JSON.

This command shows all configuration fields including prefix, id_length, etc.

Example:
  peb config`,
		Action: func(c *cli.Context) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := config.MaybeUpdatePlugin(cfg); err != nil {
				return fmt.Errorf("failed to update plugin: %w", err)
			}

			type OutputConfig struct {
				Prefix   string `json:"prefix"`
				IDLength int    `json:"id_length"`
			}

			outputCfg := OutputConfig{
				Prefix:   cfg.Prefix,
				IDLength: cfg.IDLength,
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			if err := encoder.Encode(outputCfg); err != nil {
				return fmt.Errorf("failed to encode config: %w", err)
			}

			return nil
		},
	}
}
