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
		Description: `Display the full details of one or more pebs as formatted JSON.

This command shows all peb fields including id, title, type, status,
created/changed timestamps, blocked-by list, and markdown content.

Examples:
  peb read peb-xxxx
  peb read peb-xxxx peb-yyyy peb-zzzz`,
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("at least one peb ID is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			s := store.New(cfg.PebblesDir(), cfg.Prefix)
			if err := s.Load(); err != nil {
				return err
			}

			pebIDs := c.Args().Slice()
			pebs := make([]interface{}, 0, len(pebIDs))

			for _, pebID := range pebIDs {
				p, ok := s.Get(pebID)
				if !ok {
					return fmt.Errorf("peb %s not found", pebID)
				}
				pebs = append(pebs, p)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			if len(pebs) == 1 {
				if err := encoder.Encode(pebs[0]); err != nil {
					return fmt.Errorf("failed to encode peb: %w", err)
				}
			} else {
				if err := encoder.Encode(pebs); err != nil {
					return fmt.Errorf("failed to encode pebs: %w", err)
				}
			}

			return nil
		},
	}
}
