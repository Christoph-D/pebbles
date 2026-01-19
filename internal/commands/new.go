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
		Description: `Create a new peb from JSON input via stdin.

Required fields:
  title     Short description of the peb
  content   Markdown description

Optional fields:
  type      One of: bug, feature, epic, task (default: bug)
  blocked-by Array of peb IDs this peb depends on

Examples:
  peb new <<'EOF'
  {"title":"Fix login bug","content":"Users cannot log in"}
  EOF
  
  peb new <<'EOF'
  {"title":"Add feature","content":"Details...","type":"feature"}
  EOF
  
  peb new <<'EOF'
  {"title":"Dependent task","content":"...","blocked-by":["peb-xxxx"]}
  EOF`,
		Action: func(c *cli.Context) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			s := store.New(cfg.PebblesDir(), cfg.Prefix)
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
						return fmt.Errorf("Referenced peb(s) not found: %s", extractInvalidID(err))
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

			if err := s.Save(p); err != nil {
				return fmt.Errorf("failed to save peb: %w", err)
			}

			fmt.Printf("Created new peb %s\n", id)
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
