//go:build !android

package platform

import (
	"errors"
	"os/exec"
	"runtime"
)

// ErrNoOpener means files can't be opened in another app from here; use
// the share sheet instead.
var ErrNoOpener = errors.New("opening files isn't supported here")

// CanOpen reports whether Open works for the file at path.
func CanOpen(path string) bool {
	switch runtime.GOOS {
	case "windows", "darwin", "linux", "freebsd":
		return true
	}
	return false
}

// Open opens a file in its default app (a video in the system player).
func Open(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux", "freebsd":
		cmd = exec.Command("xdg-open", path)
	default:
		return ErrNoOpener
	}
	hideWindow(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}

// Share opens the share sheet for a file in shared storage. ok is false
// when it can't, and the app's own share (which copies the file) is used.
func Share(path string) (ok bool, err error) { return false, nil }

// PublicDir is where results go on Android; elsewhere results go next to
// their sources, and it's "".
func PublicDir() string { return "" }
