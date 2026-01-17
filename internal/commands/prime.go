package commands

import (
	_ "embed"
	"os"
	"strings"
	"text/template"

	"github.com/Christoph-D/pebbles/internal/config"
	"github.com/urfave/cli/v2"
)

//go:embed prompt.md
var promptTemplate string

func PrimeCommand() *cli.Command {
	return &cli.Command{
		Name:  "prime",
		Usage: "Print the content of prompt.md",
		Action: func(c *cli.Context) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			tmpl, err := template.New("prompt").Parse(promptTemplate)
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
