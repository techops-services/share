package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy copies the given text to the system clipboard.
// It uses platform-specific commands: pbcopy (macOS), xclip or xsel (Linux),
// clip.exe (Windows).
func Copy(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return run("pbcopy", text)
	case "linux":
		return copyLinux(text)
	case "windows":
		return run("clip.exe", text)
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
}

// copyLinux tries xclip first, then xsel.
func copyLinux(text string) error {
	if _, err := exec.LookPath("xclip"); err == nil {
		return run("xclip", text, "-selection", "clipboard")
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		return run("xsel", text, "--clipboard", "--input")
	}
	return fmt.Errorf("no clipboard tool found: install xclip or xsel")
}

// run executes a command, passing text via stdin.
// The first variadic arg after the command name is the stdin text;
// remaining args are command-line arguments.
func run(name string, stdinText string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(stdinText)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard command %q failed: %w", name, err)
	}
	return nil
}
