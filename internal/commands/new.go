package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

type NewInput struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Type      string   `json:"type"`
	BlockedBy []string `json:"blocked-by"`
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "new",
		Usage: "Create a new peb",
		Action: func(c *cli.Context) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			s := store.New(cfg.PebblesDir())
			if err := s.Load(); err != nil {
				return err
			}

			var input NewInput
			decoder := json.NewDecoder(os.Stdin)
			if err := decoder.Decode(&input); err != nil {
				return fmt.Errorf("failed to parse JSON input: %w", err)
			}

			if input.Title == "" {
				return fmt.Errorf("title is required")
			}
			if input.Content == "" {
				return fmt.Errorf("content is required")
			}

			if len(input.BlockedBy) > 0 {
				if err := peb.ValidateBlockedBy(s, nil, input.BlockedBy); err != nil {
					if peb.IsInvalidReference(err) {
						return fmt.Errorf("Referenced pebble(s) not found: %s", extractInvalidID(err))
					}
					return err
				}
			}

			pebType := peb.TypeBug
			if input.Type != "" {
				pebType = peb.Type(input.Type)
			}

			id, err := s.GenerateUniqueID(cfg.Prefix, cfg.IDLength)
			if err != nil {
				return fmt.Errorf("failed to generate ID: %w", err)
			}

			p := peb.New(id, input.Title, pebType, peb.StatusNew, input.Content)
			p.BlockedBy = input.BlockedBy

			filename := peb.Filename(p)
			if err := s.Save(p); err != nil {
				return fmt.Errorf("failed to save peb: %w", err)
			}

			fmt.Printf("Created new pebble %s in .pebbles/%s\n", id, filename)
			return nil
		},
	}
}

func extractInvalidID(err error) string {
	msg := err.Error()
	if idx := len(peb.ErrInvalidReference.Error()) + 2; idx < len(msg) {
		return msg[idx:]
	}
	return ""
}
