package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

func ReadCommand() *cli.Command {
	return &cli.Command{
		Name:  "read",
		Usage: "Display peb content as JSON",
		Description: `Display the full details of a peb as formatted JSON.

This command shows all peb fields including id, title, type, status,
created/changed timestamps, blocked-by list, and markdown content.

Example:
  peb read peb-xxxx`,
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("peb ID is required")
			}

			pebID := c.Args().First()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			s := store.New(cfg.PebblesDir())
			if err := s.Load(); err != nil {
				return err
			}

			p, ok := s.Get(pebID)
			if !ok {
				return fmt.Errorf("peb %s not found", pebID)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(p); err != nil {
				return fmt.Errorf("failed to encode peb: %w", err)
			}

			return nil
		},
	}
}
