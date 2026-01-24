package commands

import (
	"fmt"
	"strings"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

func DeleteCommand() *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Delete one or more pebs",
		Description: `Delete one or more pebs by their IDs.

This command permanently removes peb files from storage. The deleted pebs
cannot be recovered.

Examples:
  peb delete peb-xxxx
  peb delete peb-xxxx peb-yyyy peb-zzzz`,
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("at least one peb ID is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := config.MaybeUpdatePlugin(cfg); err != nil {
				return fmt.Errorf("failed to update plugin: %w", err)
			}

			s := store.New(cfg.PebblesDir(), cfg.Prefix)
			if err := s.Load(); err != nil {
				return err
			}

			pebIDs := c.Args().Slice()

			for _, pebID := range pebIDs {
				if _, ok := s.Get(pebID); !ok {
					return fmt.Errorf("peb %s not found", pebID)
				}
			}

			pebsToDelete := make([]*peb.Peb, 0, len(pebIDs))
			for _, pebID := range pebIDs {
				p, _ := s.Get(pebID)
				pebsToDelete = append(pebsToDelete, p)
			}

			for _, p := range pebsToDelete {
				if err := s.Delete(p); err != nil {
					return fmt.Errorf("failed to delete peb %s: %w", p.ID, err)
				}
			}

			if len(pebIDs) == 1 {
				fmt.Printf("Deleted peb %s.\n", pebIDs[0])
			} else {
				fmt.Printf("Deleted pebs %s.\n", strings.Join(pebIDs, " "))
			}

			return nil
		},
	}
}
