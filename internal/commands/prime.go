package commands

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/urfave/cli/v2"
)

//go:embed data/prompt.md
var promptTemplate string

//go:embed data/prompt-mcp.md
var mcpPromptTemplate string

func PrimeCommand() *cli.Command {
	return &cli.Command{
		Name:  "prime",
		Usage: "Primes the coding agent",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "mcp",
				Usage: "Print the MCP prompt instead of the normal prompt",
			},
		},
		Action: func(c *cli.Context) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := config.MaybeUpdatePlugin(cfg); err != nil {
				return fmt.Errorf("failed to update plugin: %w", err)
			}

			templateContent := promptTemplate
			if c.Bool("mcp") {
				templateContent = mcpPromptTemplate
			}

			tmpl, err := template.New("prompt").Parse(templateContent)
			if err != nil {
				return err
			}

			data := struct {
				PebbleIDSuffix   string
				PebbleIDPattern  string
				PebbleIDPattern2 string
				PebbleIDPattern3 string
			}{
				PebbleIDSuffix:   strings.Repeat("x", cfg.IDLength),
				PebbleIDPattern:  cfg.Prefix + "-" + strings.Repeat("x", cfg.IDLength),
				PebbleIDPattern2: cfg.Prefix + "-" + strings.Repeat("y", cfg.IDLength),
				PebbleIDPattern3: cfg.Prefix + "-" + strings.Repeat("z", cfg.IDLength),
			}

			return tmpl.Execute(os.Stdout, data)
		},
	}
}
