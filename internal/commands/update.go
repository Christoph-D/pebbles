package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

type UpdateInput struct {
	Title     *string   `json:"title"`
	Content   *string   `json:"content"`
	Type      *string   `json:"type"`
	Status    *string   `json:"status"`
	BlockedBy *[]string `json:"blocked-by,omitempty"`
}

const maxOutputLength = 100

func truncate(s string) string {
	if len(s) > maxOutputLength {
		return s[:maxOutputLength] + "..."
	}
	return s
}

func UpdateCommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update peb fields",
		Description: `Update an existing peb. Takes a peb ID as argument and JSON input via stdin or as a CLI argument.

All fields are optional. Only provided fields will be updated.

Available fields:
  title      Short description (use stdin to avoid quoting issues)
  content    Markdown description (use stdin to avoid quoting issues)
  type       One of: bug, feature, epic, task
  status     One of: new, in-progress, fixed, wont-fix
  blocked-by Array of peb IDs this peb depends on

Examples:
  peb update peb-xxxx '{"status":"in-progress"}'
  peb update peb-xxxx '{"type":"feature"}'
  peb update peb-xxxx '{"blocked-by":["peb-yyyy","peb-zzzz"]}'
  
  peb update peb-xxxx <<'EOF'
  {"title":"New title"}
  EOF
  
  peb update peb-xxxx <<'EOF'
  {"content":"Detailed description\n\nWith multiple paragraphs"}
  EOF`,
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("peb ID is required")
			}

			pebID := c.Args().First()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := config.MaybeUpdatePlugin(); err != nil {
				return fmt.Errorf("failed to update plugin: %w", err)
			}

			s := store.New(cfg.PebblesDir(), cfg.Prefix)
			if err := s.Load(); err != nil {
				return err
			}

			p, ok := s.Get(pebID)
			if !ok {
				return fmt.Errorf("peb %s not found", pebID)
			}

			var input UpdateInput

			if c.NArg() > 1 {
				args := c.Args().Tail()
				lastArg := args[len(args)-1]
				if strings.HasPrefix(lastArg, "{") || strings.HasPrefix(lastArg, "[") {
					jsonStr := lastArg
					if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
						return fmt.Errorf("failed to parse JSON input: %w", err)
					}
				} else {
					decoder := json.NewDecoder(os.Stdin)
					if err := decoder.Decode(&input); err != nil {
						return fmt.Errorf("failed to parse JSON input: %w", err)
					}
				}
			} else {
				decoder := json.NewDecoder(os.Stdin)
				if err := decoder.Decode(&input); err != nil {
					return fmt.Errorf("failed to parse JSON input: %w", err)
				}
			}

			oldFilename := peb.Filename(p)
			oldTitle := p.Title
			oldType := p.Type
			oldStatus := p.Status

			if input.Title != nil {
				p.Title = *input.Title
			}
			if input.Content != nil {
				p.Content = *input.Content
			}
			if input.Type != nil {
				p.Type = peb.Type(*input.Type)
			}
			if input.Status != nil {
				p.Status = peb.Status(*input.Status)
			}
			if input.BlockedBy != nil {
				p.BlockedBy = *input.BlockedBy
			}

			if input.BlockedBy != nil && len(*input.BlockedBy) > 0 {
				if err := peb.ValidateBlockedBy(s, p, *input.BlockedBy); err != nil {
					if peb.IsInvalidReference(err) {
						return fmt.Errorf("Referenced peb(s) not found: %s", extractInvalidID(err))
					}
					return err
				}
				if err := peb.CheckCycle(s, p.ID, *input.BlockedBy); err != nil {
					return err
				}
			}

			p.UpdateTimestamp()

			if input.Title != nil {
				oldPath := filepath.Join(cfg.PebblesDir(), oldFilename)
				newFilename := peb.Filename(p)
				newPath := filepath.Join(cfg.PebblesDir(), newFilename)
				if err := os.Rename(oldPath, newPath); err != nil {
					return fmt.Errorf("failed to rename file: %w", err)
				}
			}

			if err := s.Save(p); err != nil {
				return fmt.Errorf("failed to save peb: %w", err)
			}

			if input.Status != nil && oldStatus != p.Status {
				fmt.Printf("Updated status of %s to %s.\n", pebID, p.Status)
			}
			if input.Title != nil && oldTitle != p.Title {
				fmt.Printf("Updated title of %s to %q.\n", pebID, truncate(p.Title))
			}
			if input.Content != nil && p.Content != "" {
				fmt.Printf("Updated content of %s to %q.\n", pebID, truncate(p.Content))
			}
			if input.Type != nil && oldType != p.Type {
				fmt.Printf("Updated type of %s to %s.\n", pebID, p.Type)
			}
			if input.BlockedBy != nil {
				if len(*input.BlockedBy) > 0 {
					fmt.Printf("Updated blocked-by list of %s to %v.\n", pebID, p.BlockedBy)
				} else {
					fmt.Printf("Cleared blocked-by list of %s.\n", pebID)
				}
			}

			return nil
		},
	}
}
