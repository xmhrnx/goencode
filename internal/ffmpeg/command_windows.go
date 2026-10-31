//go:build windows

package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const createNoWindow = 0x08000000

// command starts a program without flashing a console window. Canceling ctx
// ends the program and every process it started: package-manager launchers
// run the real ffmpeg as a child, which would otherwise keep encoding.
func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	cmd.Cancel = func() error {
		kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
		kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
		if kill.Run() != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
	// Don't wait forever for output pipes a stray child still holds.
	cmd.WaitDelay = 3 * time.Second
	return cmd
}

var scoopShimPath = regexp.MustCompile(`(?m)^\s*path\s*=\s*"?([^"\r\n]+)"?`)

// realProgram returns the program a package manager's launcher stands for:
// Chocolatey's bin\ffmpeg.exe and Scoop's shims\ffmpeg.exe are small
// programs that start the real one. Running it directly shows ffmpeg's
// own path and makes stopping an encode immediate.
func realProgram(path string) string {
	dir, base := filepath.Split(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Scoop: ffmpeg.shim next to the launcher names the target.
	if data, err := os.ReadFile(filepath.Join(dir, name+".shim")); err == nil {
		if m := scoopShimPath.FindSubmatch(data); m != nil && isFile(string(m[1])) {
			return string(m[1])
		}
	}

	// Chocolatey: <root>\bin\ffmpeg.exe runs <root>\lib\ffmpeg*\tools\...\ffmpeg.exe.
	if strings.EqualFold(filepath.Base(filepath.Clean(dir)), "bin") {
		root := filepath.Dir(filepath.Clean(dir))
		if _, err := os.Stat(filepath.Join(root, "lib")); err == nil && strings.Contains(strings.ToLower(root), "chocolatey") {
			for _, pattern := range []string{
				filepath.Join(root, "lib", "ffmpeg*", "tools", "*", "bin", base),
				filepath.Join(root, "lib", "ffmpeg*", "tools", "bin", base),
				filepath.Join(root, "lib", "ffmpeg*", "tools", base),
			} {
				if found, _ := filepath.Glob(pattern); len(found) > 0 {
					return found[0]
				}
			}
		}
	}
	return path
}
