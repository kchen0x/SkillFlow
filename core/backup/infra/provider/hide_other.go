//go:build !windows

package provider

import "os/exec"

func hideConsole(cmd *exec.Cmd) {}
