package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/Christoph-D/pebbles/internal/peb"
	"github.com/Christoph-D/pebbles/internal/store"
	"github.com/urfave/cli/v2"
)

type filterFunc func(*peb.Peb) bool

func QueryCommand() *cli.Command {
	return &cli.Command{
		Name:  "query",
		Usage: "Search and list pebs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "fields",
				Usage:   "Comma-separated list of fields to output (default: id,type,status,title)",
				Value:   "id,type,status,title",
				Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			s := store.New(cfg.PebblesDir())
			if err := s.Load(); err != nil {
				return err
			}

			filters, err := parseFilters(c.Args().Slice())
			if err != nil {
				return err
			}

			fields, err := parseFields(c.String("fields"))
			if err != nil {
				return err
			}

			pebs := s.All()

			for _, p := range pebs {
				if applyFilters(p, filters) {
					output := buildOutput(p, fields)
					if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
						return fmt.Errorf("failed to encode peb: %w", err)
					}
				}
			}

			return nil
		},
	}
}

func parseFilters(args []string) ([]filterFunc, error) {
	var filters []filterFunc

	for _, arg := range args {
		if !strings.Contains(arg, ":") {
			return nil, fmt.Errorf("invalid filter format: %s (expected key:value)", arg)
		}

		parts := strings.SplitN(arg, ":", 2)
		key, value := parts[0], parts[1]

		switch key {
		case "status":
			switch value {
			case "open":
				filters = append(filters, func(p *peb.Peb) bool {
					for _, s := range peb.StatusOpen {
						if p.Status == s {
							return true
						}
					}
					return false
				})
			case "closed":
				filters = append(filters, func(p *peb.Peb) bool {
					for _, s := range peb.StatusClosed {
						if p.Status == s {
							return true
						}
					}
					return false
				})
			default:
				filters = append(filters, func(p *peb.Peb) bool {
					return string(p.Status) == value
				})
			}
		case "type":
			filters = append(filters, func(p *peb.Peb) bool {
				return string(p.Type) == value
			})
		case "blocked-by":
			filters = append(filters, func(p *peb.Peb) bool {
				for _, id := range p.BlockedBy {
					if id == value {
						return true
					}
				}
				return false
			})
		default:
			return nil, fmt.Errorf("unknown filter key: %s", key)
		}
	}

	return filters, nil
}

func applyFilters(p *peb.Peb, filters []filterFunc) bool {
	for _, f := range filters {
		if !f(p) {
			return false
		}
	}
	return true
}

func parseFields(fieldsStr string) ([]string, error) {
	fields := strings.Split(fieldsStr, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		switch field {
		case "id", "type", "status", "title", "created", "changed", "blocked-by":
		default:
			return nil, fmt.Errorf("unknown field: %s", field)
		}
	}
	return fields, nil
}

func buildOutput(p *peb.Peb, fields []string) map[string]interface{} {
	output := make(map[string]interface{})
	for _, field := range fields {
		switch field {
		case "id":
			output["id"] = p.ID
		case "type":
			output["type"] = p.Type
		case "status":
			output["status"] = p.Status
		case "title":
			output["title"] = p.Title
		case "created":
			output["created"] = p.Created
		case "changed":
			output["changed"] = p.Changed
		case "blocked-by":
			output["blocked-by"] = p.BlockedBy
		}
	}
	return output
}
