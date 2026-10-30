//go:build windows

package ffmpeg

import (
	"context"
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// command starts a program without flashing a console window.
func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	return cmd
}
