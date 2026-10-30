//go:build !android && !ios

package ffmpeg

import "runtime"

const execSupported = true

// platformDirs lists where package managers put ffmpeg when it may not be on
// the PATH a GUI app inherits (apps started from the macOS Dock get a
// minimal PATH without Homebrew's folder).
func platformDirs() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/opt/homebrew/bin", "/usr/local/bin", "/opt/local/bin"}
	case "linux", "freebsd":
		return []string{"/usr/bin", "/usr/local/bin", "/snap/bin"}
	case "windows":
		return []string{`C:\ffmpeg\bin`, `C:\Program Files\ffmpeg\bin`, `C:\ProgramData\chocolatey\bin`}
	}
	return nil
}
