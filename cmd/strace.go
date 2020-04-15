package cmd

import (
	"os/exec"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var straceCommand = &cli.Command{
	Name:   "strace",
	Usage:  "Run strace of a program e.g. strace ps aux",
	Action: strace,
}

func strace(c *cli.Context) error {

	// Pins the goroutine to the thread
	// which is necessary for ptrace to work correctly
	// (probably not needed) in non-concurrent context though,
	//  but still putting it here as good habit)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// First, define the child process and set
	// the ptrace flag to true. This essentially:
	// - forks the current process
	// - calls ptrace with PTRACE_TRACEME in the child
	// - stops the execution of the child
	cmd := exec.Command(c.Args().First(), c.Args().Slice()[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}
	log.WithFields(log.Fields{
		"path":    cmd.Path,
		"args":    cmd.Args,
		"cliArgs": c.Args().Slice()[1:],
	}).Info("Running command")
	cmd.Start()

	// Get the child pid so that we can work with it
	childPid := cmd.Process.Pid

	// Wait for the child to stop itself after calling ptrace
	_, err := waitPid(childPid)
	if err != nil {
		log.WithField("pid", childPid).Error("Error while waiting for child process")
		return err
	}

	// Now, we will use ptrace and waitpid from parent to:
	// - tell the child to run until it either triggers syscall or comes back up after (see man for PTRACE_SYSCALL)
	// - wait until a trap is received
	// - on every second trap received (i.e. only on traps triggered AFTER the syscalls), we will print the value of Orig_rax register

	traps := 0

	for {
		// Call ptrace to tell the child to run until the next syscall invocation or coming back from it
		err := runUntilSyscallStartOrEnd(childPid)
		if err != nil {
			log.WithField("pid", childPid).Error("Errow while instructing child to run until the next syscall")
			return err
		}

		// Wait for the child to send us a trap when it does syscalls
		waitStatus, err := waitPid(childPid)
		if err != nil {
			log.WithField("pid", childPid).Error("Error while waiting for child process")
			return err
		}

		// If the child is done, then we are also done
		if waitStatus.Exited() {
			log.WithField("status", waitStatus.ExitStatus()).Info("Child exited")
			return nil
		}

		// Print register value on every second trap so that we print only after the syscall is complete
		if traps > 0 && traps%2 == 0 {
			origRax, err := readOrigRax(childPid)
			if err != nil {
				log.WithField("pid", childPid).Error("Error while reading Orig_rax for child process")
				return err
			}
			log.WithField("orig_rax", origRax).Info("Got value of Orig_rax (syscall id)")
		}

		traps++
	}
}

func waitPid(pid int) (syscall.WaitStatus, error) {
	var waitStatus syscall.WaitStatus
	_, err := syscall.Wait4(pid, &waitStatus, 0, nil)
	if err != nil {
		return waitStatus, err
	}
	return waitStatus, nil
}

func readOrigRax(pid int) (uint64, error) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		return 0, err
	}
	return regs.Orig_rax, nil
}

func runUntilSyscallStartOrEnd(pid int) error {
	err := syscall.PtraceSyscall(pid, 0)
	if err != nil {
		return err
	}
	return nil
}
