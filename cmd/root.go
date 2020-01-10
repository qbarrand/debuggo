package cmd

import (
	"github.com/urfave/cli/v2"
)

func CLI() *cli.App {
	const name = "debuggo"

	app := cli.NewApp()

	app.Name = name
	app.Version = "0.0.0"
	app.Authors = []*cli.Author{
		{Name: "Quentin Barrand", Email: "quentin@quba.fr"},
		{Name: "Adam Krajewski"},
	}

	app.UsageText = name + " PATH [ARGS...]"

	app.Action = func(c *cli.Context) error {
		if c.NArg() < 1 {
			// at least one path must be specified
			return cli.ShowAppHelp(c)
		}

		return nil
	}

	return app
}
