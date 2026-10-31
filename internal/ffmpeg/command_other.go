//go:build !windows

package ffmpeg

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

// command runs a program in its own process group, so canceling ctx ends it
// and anything it started (wrapper scripts that run the real ffmpeg).
func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
	// Don't wait forever for output pipes a stray child still holds.
	cmd.WaitDelay = 3 * time.Second
	return cmd
}

// realProgram returns path: package managers here link to the real program.
func realProgram(path string) string { return path }
