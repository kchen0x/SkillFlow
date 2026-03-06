//go:build !windows

package backup

import "os/exec"

func hideConsole(cmd *exec.Cmd) {}
