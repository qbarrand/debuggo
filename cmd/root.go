package cmd

import (
	"github.com/urfave/cli/v2"
)

func CLI() *cli.App {
	app := cli.NewApp()
	app.Version = "0.0.0"
	app.Authors = []*cli.Author{
		{Name: "Quentin Barrand", Email: "quentin@quba.fr"},
		{Name: "Adam Krajewski"},
	}
	app.Commands = []*cli.Command{
		straceCommand,
		debugCommand,
	}

	return app
}
