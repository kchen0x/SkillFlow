package editor

import (
	"fmt"
	"os/exec"
	goruntime "runtime"
)

// OpenFile opens the given file path in the system default editor.
// On macOS: uses `open <path>`
// On Linux: uses `xdg-open <path>`
// On Windows: uses `start <path>` via cmd.exe
func OpenFile(path string) error {
	switch goruntime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", goruntime.GOOS)
	}
}
