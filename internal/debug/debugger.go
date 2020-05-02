package debug

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const prompt = "dbg> "

var int3 = []byte{0xCC}

type cmdHandler func(c *cmdContext) error

type cmdContext struct {
	Cmd  string
	Args []string
}

type Debugger interface {
	Debug(execPath string) error
}

type debugger struct {
	// Will be further populated with data as we implement more features
	pid                  int
	breakpointToOriginal map[uintptr]byte

	// Store commands both in slice and map to avoid
	// iterating ove map each time we need the commands
	// e.g. to display help
	cmdToHandler map[string]cmdHandler
}

func NewDebugger() Debugger {
	d := &debugger{
		cmdToHandler:         make(map[string]cmdHandler),
		breakpointToOriginal: make(map[uintptr]byte),
	}

	// There are also special 'exit' and 'quit' commands
	// which we act on in Loop
	d.registerCmd("continue", d.cont)
	d.registerCmd("step", d.step)
	d.registerCmd("breakpoint", d.setBreakpoint)
	d.registerCmd("show", d.show)

	return d
}

func (d *debugger) registerCmd(name string, handler cmdHandler) {
	d.cmdToHandler[name] = handler
}

func (d *debugger) Debug(execPath string) error {

	pid, err := runTracee(execPath)
	if err != nil {
		return err
	}
	d.pid = pid

	fmt.Printf("Debugging started for %s, pid %d\n", execPath, pid)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(prompt)
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

// Continues execution of the tracee, also after breakpoints
func (d *debugger) cont(c *cmdContext) error {
	var regs syscall.PtraceRegs
	if err := syscall.PtraceGetRegs(d.pid, &regs); err != nil {
		return err
	}

	// First, we need to check if we stopped at the breakpoint
	// If yes, the instruction pointer will be just after the breakpoint address
	breakpoint := uintptr(regs.PC() - 1)
	_, ok := d.breakpointToOriginal[breakpoint]
	if ok {
		// We stopped because of a breakpoint
		// First, let's remove the breakpoint
		original := d.breakpointToOriginal[breakpoint]
		if _, err := syscall.PtracePokeText(d.pid, breakpoint, []byte{original}); err != nil {
			return err
		}

		// Then, let's move the instruction pointer back
		// so that we execute the instruction where we set up the breakpoint
		regs.SetPC(regs.PC() - 1)
		if err := syscall.PtraceSetRegs(d.pid, &regs); err != nil {
			return err
		}

		// Make a single step and wait for the process to stop execution
		// so that we can put the breakpoint in place again
		if err := syscall.PtraceSingleStep(d.pid); err != nil {
			return err
		}
		if _, err := syscall.Wait4(d.pid, nil, 0, nil); err != nil {
			return err
		}

		// Finally, restore the breakpoint
		if _, err := syscall.PtracePokeText(d.pid, breakpoint, int3); err != nil {
			return err
		}
	}

	// Just continue
	if err := syscall.PtraceCont(d.pid, 0); err != nil {
		return err
	}

	return nil
}

// Makes a single step
func (d *debugger) step(c *cmdContext) error {
	if err := syscall.PtraceSingleStep(d.pid); err != nil {
		return err
	}

	return nil
}

// Sets breakpoint at the given address
func (d *debugger) setBreakpoint(c *cmdContext) error {
	if len(c.Args) != 1 {
		fmt.Println("usage: breakpoint <hexaddr>")
		return nil
	}

	addrStr := c.Args[0]
	addrUint64, err := hexStringToUint64(addrStr)
	if err != nil {
		return err
	}

	breakpoint := uintptr(addrUint64)
	fmt.Printf("Setting breakpoint at 0x%x \n", breakpoint)

	buf := make([]byte, 1)
	_, err = syscall.PtracePeekText(d.pid, breakpoint, buf)
	if err != nil {
		return err
	}

	d.breakpointToOriginal[breakpoint] = buf[0]

	if _, err := syscall.PtracePokeText(d.pid, breakpoint, int3); err != nil {
		return err
	}

	return nil
}

// Displays various data from the debugger
func (d *debugger) show(c *cmdContext) error {
	if len(c.Args) != 1 {
		fmt.Println("Usage: show pc|breakpoints")
	}

	what := c.Args[0]
	switch what {

	case "pc":
		var regs syscall.PtraceRegs
		if err := syscall.PtraceGetRegs(d.pid, &regs); err != nil {
			return err
		}

		fmt.Printf("PC at 0x%x\n", regs.PC())

	case "breakpoints":
		if len(d.breakpointToOriginal) == 0 {
			fmt.Println("No breakpoints set")
			break
		}

		fmt.Println("Breakpoints set:")
		for breakpoint, _ := range d.breakpointToOriginal {
			fmt.Printf("  - 0x%x\n", breakpoint)
		}
	}

	return nil
}

// Starts the traced process i.e. tracee
func runTracee(execPath string) (int, error) {
	cmd := exec.Command(execPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	pid := cmd.Process.Pid

	if _, err := syscall.Wait4(pid, nil, 0, nil); err != nil {
		return 0, err
	}

	return pid, nil
}

// Converts hex string to uint64
func hexStringToUint64(s string) (uint64, error) {
	// Remove leading 0x if present
	s = strings.Replace(s, "0x", "", 1)

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}

	for len(bytes) < 8 {
		bytes = append([]byte{0}, bytes...)
	}

	return binary.BigEndian.Uint64(bytes), nil
}
