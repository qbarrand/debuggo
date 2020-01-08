package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/qbarrand/debuggo/cmd"
)

func main() {
	app := cmd.CLI()
	if app == nil {
		logrus.Fatalf("Could not get the command-line parser")
		return
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatalf("General error caught")
	}
}