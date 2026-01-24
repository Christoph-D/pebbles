package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
		Description: `Query pebs using filters. Multiple filters are combined with AND logic.

ID filters:
  id:peb-xxxx          Show peb with specific ID
  id:(peb-xxxx|peb-yyyy)  Show pebs with any of the listed IDs

Status filters:
  status:new           Show new pebs
  status:in-progress   Show in-progress pebs
  status:fixed         Show fixed pebs
  status:wont-fix      Show pebs marked as wont-fix
  status:open          Show all open pebs (new or in-progress)
  status:closed        Show all closed pebs (fixed or wont-fix)

Type filters:
  type:bug             Show bugs
  type:feature         Show features
  type:epic            Show epics
  type:task            Show tasks

OR syntax:
  id:(peb-xxxx|peb-yyyy)    Show pebs with any of the listed IDs
  type:(bug|feature)        Show bugs or features
  status:(new|fixed)        Show new or fixed pebs

Other filters:
  blocked-by:peb-xxxx  Show pebs blocked by a specific peb ID

Examples:
  peb query                          List all pebs
  peb query id:peb-xxxx              Show peb-xxxx
  peb query id:(peb-xxxx|peb-yyyy)   Show peb-xxxx or peb-yyyy
  peb query status:new               Show all new pebs
  peb query type:bug                 Show all bugs
  peb query status:new type:feature  Show new features only
  peb query type:(bug|feature)       Show bugs or features
  peb query blocked-by:peb-xxxx      Show pebs blocked by peb-xxxx
  peb query --fields id,title        Show only id and title fields`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "fields",
				Usage:   "Comma-separated list of fields to output (default: id,type,status,title,blocked-by)",
				Value:   "id,type,status,title,blocked-by",
				Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
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

			filters, err := parseFilters(c.Args().Slice())
			if err != nil {
				return err
			}

			fields, err := parseFields(c.String("fields"))
			if err != nil {
				return err
			}

			pebs := s.All()

			sort.Slice(pebs, func(i, j int) bool {
				return pebs[i].ID < pebs[j].ID
			})

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
			if values := parseOrValues(value); len(values) > 0 {
				filters = append(filters, func(p *peb.Peb) bool {
					for _, v := range values {
						switch v {
						case "open":
							if peb.IsOpen(p.Status) {
								return true
							}
						case "closed":
							if peb.IsClosed(p.Status) {
								return true
							}
						default:
							if string(p.Status) == v {
								return true
							}
						}
					}
					return false
				})
			} else {
				switch value {
				case "open":
					filters = append(filters, func(p *peb.Peb) bool {
						return peb.IsOpen(p.Status)
					})
				case "closed":
					filters = append(filters, func(p *peb.Peb) bool {
						return peb.IsClosed(p.Status)
					})
				default:
					filters = append(filters, func(p *peb.Peb) bool {
						return string(p.Status) == value
					})
				}
			}
		case "type":
			if values := parseOrValues(value); len(values) > 0 {
				filters = append(filters, func(p *peb.Peb) bool {
					for _, v := range values {
						if string(p.Type) == v {
							return true
						}
					}
					return false
				})
			} else {
				filters = append(filters, func(p *peb.Peb) bool {
					return string(p.Type) == value
				})
			}
		case "id":
			if values := parseOrValues(value); len(values) > 0 {
				filters = append(filters, func(p *peb.Peb) bool {
					for _, v := range values {
						if p.ID == v {
							return true
						}
					}
					return false
				})
			} else {
				filters = append(filters, func(p *peb.Peb) bool {
					return p.ID == value
				})
			}
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

func parseOrValues(value string) []string {
	if !strings.HasPrefix(value, "(") || !strings.HasSuffix(value, ")") {
		return nil
	}

	inner := strings.Trim(value[1:len(value)-1], " ")
	if !strings.Contains(inner, "|") {
		trimmed := strings.TrimSpace(inner)
		if trimmed != "" {
			return []string{trimmed}
		}
		return nil
	}

	parts := strings.Split(inner, "|")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	return values
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
	var parsedFields []string
	hasID := false
	for _, field := range fields {
		field = strings.TrimSpace(field)
		switch field {
		case "id", "type", "status", "title", "created", "changed", "blocked-by":
			parsedFields = append(parsedFields, field)
			if field == "id" {
				hasID = true
			}
		default:
			return nil, fmt.Errorf("unknown field: %s", field)
		}
	}
	if !hasID {
		parsedFields = append([]string{"id"}, parsedFields...)
	}
	return parsedFields, nil
}

func buildOutput(p *peb.Peb, fields []string) *peb.PebJSON {
	output := &peb.PebJSON{
		ID: p.ID,
	}
	for _, field := range fields {
		switch field {
		case "type":
			output.Type = p.Type
		case "status":
			output.Status = p.Status
		case "title":
			output.Title = p.Title
		case "created":
			output.Created = p.Created
		case "changed":
			output.Changed = p.Changed
		case "blocked-by":
			if len(p.BlockedBy) > 0 {
				output.BlockedBy = p.BlockedBy
			}
		}
	}
	return output
}
