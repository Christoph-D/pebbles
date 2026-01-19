package commands

import (
	"fmt"
	"strings"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

func buildDependencyMap(allPebs []*peb.Peb) map[string][]string {
	depMap := make(map[string][]string)
	for _, p := range allPebs {
		for _, blockedBy := range p.BlockedBy {
			depMap[blockedBy] = append(depMap[blockedBy], p.ID)
		}
	}
	return depMap
}

func DeleteCommand() *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Delete one or more pebs",
		Description: `Delete one or more pebs by their IDs.

This command permanently removes peb files from storage. The deleted pebs
cannot be recovered.

A peb with dependants can only be deleted if all of its dependants are
also being deleted in the same command.

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

			allPebs := s.All()
			depMap := buildDependencyMap(allPebs)

			pebIDSet := make(map[string]bool)
			for _, pebID := range pebIDs {
				pebIDSet[pebID] = true
			}

			for _, pebID := range pebIDs {
				if dependents, exists := depMap[pebID]; exists && len(dependents) > 0 {
					var notBeingDeleted []string
					for _, depID := range dependents {
						if !pebIDSet[depID] {
							notBeingDeleted = append(notBeingDeleted, depID)
						}
					}
					if len(notBeingDeleted) > 0 {
						return fmt.Errorf("cannot delete %s: referenced by blocked-by in peb(s) not being deleted: %s", pebID, strings.Join(notBeingDeleted, ", "))
					}
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
