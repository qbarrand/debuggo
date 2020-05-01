package cmd

import (
	"github.com/qbarrand/debuggo/internal/debug"
	"github.com/urfave/cli/v2"
)

var debugCommand = &cli.Command{
	Name:   "debug",
	Usage:  "Debug a program in a manner similar to gdb",
	Action: dbg,
}

func dbg(c *cli.Context) error {
	return debug.NewDebugger().Loop()
}
