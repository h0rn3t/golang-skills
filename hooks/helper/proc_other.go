//go:build !unix

package main

import (
	"os"
	"os/exec"
)

func ownGroup(*exec.Cmd) {}

func killGroup(cmd *exec.Cmd) {
	_ = cmd.Process.Kill() // the process may be gone already
}

func openInput(p string) (*os.File, error) {
	return os.Open(p)
}

func exitStatus(ps *os.ProcessState) int {
	return ps.ExitCode()
}
