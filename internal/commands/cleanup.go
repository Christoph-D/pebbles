package commands

import (
	"fmt"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

func CleanupCommand() *cli.Command {
	return &cli.Command{
		Name:  "cleanup",
		Usage: "Delete all closed pebs",
		Description: `Delete all pebs with status "fixed" or "wont-fix". 
This permanently removes closed pebs from the system.

Examples:
  peb cleanup  # Delete all fixed and wont-fix pebs`,
		Action: func(c *cli.Context) error {
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

			allPebs := s.All()
			deletedCount := 0

			for _, p := range allPebs {
				if peb.IsClosed(p.Status) {
					if err := s.Delete(p); err != nil {
						return fmt.Errorf("failed to delete peb %s: %w", p.ID, err)
					}
					deletedCount++
				}
			}

			if deletedCount > 0 {
				fmt.Printf("Deleted %d closed peb(s).\n", deletedCount)
			} else {
				fmt.Println("No closed pebs found to delete.")
			}

			return nil
		},
	}
}
