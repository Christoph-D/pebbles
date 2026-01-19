package main

import (
	"fmt"
	"os"

	"github.com/Christoph-D/pebbles/internal/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "peb",
		Usage: "Task tracking CLI tool",
		Commands: []*cli.Command{
			commands.InitCommand(),
			commands.NewCommand(),
			commands.ReadCommand(),
			commands.UpdateCommand(),
			commands.DeleteCommand(),
			commands.QueryCommand(),
			commands.CleanupCommand(),
			commands.PrimeCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
