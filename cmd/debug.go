package cmd

import (
	"github.com/qbarrand/debuggo/internal/debug"
	"github.com/urfave/cli/v2"
)

var debugCommand = &cli.Command{
	Name:   "debug",
	Usage:  "Debug a program in a manner similar to gdb",
	Action: db,
}

func db(c *cli.Context) error {
	d := debug.NewDebugger()
	return d.Loop()
}
