//go:build unix

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// ownGroup starts cmd in a session of its own, so killGroup reaches the
// processes it starts as well.
func ownGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func killGroup(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) // the group may be gone already
}

// openInput opens a file without blocking on a FIFO.
func openInput(p string) (*os.File, error) {
	return os.OpenFile(p, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}

// exitStatus is the exit code of a finished command, or minus the signal
// number when a signal ended it.
func exitStatus(ps *os.ProcessState) int {
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return -int(ws.Signal())
	}
	return ps.ExitCode()
}
