package debug

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const PROMPT string = "dbg> "

type cmdHandler func(c *cmdContext) error

type cmdContext struct {
	Cmd  string
	Args []string
}

type Debugger interface {
	Loop() error
}

type debugger struct {
	// Will be further populated with data as we implement more features

	// Store commands both in slice and map to avoid
	// iterating ove map each time we need the commands
	// e.g. to display help
	cmds         []string
	cmdToHandler map[string]cmdHandler
}

func NewDebugger() Debugger {
	d := &debugger{
		cmdToHandler: make(map[string]cmdHandler),
	}

	// There are also special 'exit' and 'quit' commands
	// which we act on in Loop
	d.registerCmd("continue", d.cont)
	d.registerCmd("step", d.step)
	d.registerCmd("breakpoint", d.breakpoint)

	return d
}

func (d *debugger) registerCmd(name string, handler cmdHandler) {
	d.cmds = append(d.cmds, name)
	d.cmdToHandler[name] = handler
}

func (d *debugger) Loop() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()

	for {
		fmt.Printf(PROMPT)
		line, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		split := strings.Split(string(line), " ")

		cmd, args := split[0], split[1:]
		if cmd == "exit" || cmd == "quit" {
			// Just exit the loop
			return nil
		}

		if err = d.runCommand(cmd, args); err != nil {
			return err
		}

	}
}

func (d *debugger) runCommand(cmd string, args []string) error {
	c := &cmdContext{
		Cmd:  cmd,
		Args: args,
	}

	handler, ok := d.cmdToHandler[cmd]
	if !ok {
		fmt.Println("Unknown command. Supported commands are: ")
		return nil
	}

	return handler(c)
}

func (d *debugger) cont(c *cmdContext) error {
	fmt.Println("Continuing....")
	return nil
}

func (d *debugger) step(c *cmdContext) error {
	fmt.Println("Stepping...")
	return nil
}

func (d *debugger) breakpoint(c *cmdContext) error {
	fmt.Printf("breakpoint with args %v\n", c)
	return nil
}
