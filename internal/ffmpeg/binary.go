// Package ffmpeg runs the ffmpeg and ffprobe programs: finding them, asking
// what they support, probing media files and running encodes with progress.
package ffmpeg

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrNotFound means no usable ffmpeg was found.
var ErrNotFound = errors.New("ffmpeg not found")

// ErrUnsupported means this platform can't run external programs (iOS).
var ErrUnsupported = errors.New("running ffmpeg isn't supported on this platform")

// Binaries are the resolved paths of the two programs GoEncode drives.
type Binaries struct {
	FFmpeg  string
	FFprobe string
	Version string
}

// Locate finds ffmpeg and ffprobe. A non-empty override (the path set in
// preferences, either to the program or to its folder) is tried first, then
// the folder of the GoEncode executable (bundled builds), the platform's own
// locations (an Android app's native library folder) and finally PATH.
func Locate(ctx context.Context, override string) (Binaries, error) {
	if !execSupported {
		return Binaries{}, ErrUnsupported
	}
	for _, dir := range candidateDirs(override) {
		ff, fp := programPath(dir, "ffmpeg"), programPath(dir, "ffprobe")
		if isFile(ff) && isFile(fp) {
			return withVersion(ctx, Binaries{FFmpeg: ff, FFprobe: fp})
		}
	}
	ff, err1 := exec.LookPath("ffmpeg")
	fp, err2 := exec.LookPath("ffprobe")
	if err1 != nil || err2 != nil {
		return Binaries{}, ErrNotFound
	}
	return withVersion(ctx, Binaries{FFmpeg: ff, FFprobe: fp})
}

func candidateDirs(override string) []string {
	var dirs []string
	if override = strings.TrimSpace(override); override != "" {
		if fi, err := os.Stat(override); err == nil && !fi.IsDir() {
			override = filepath.Dir(override)
		}
		dirs = append(dirs, override)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		dirs = append(dirs, dir, filepath.Join(dir, "ffmpeg"), filepath.Join(dir, "..", "Resources"))
	}
	return append(dirs, platformDirs()...)
}

// programPath names a program in dir. On Android the binaries ship as
// libffmpeg.so and libffprobe.so, the only files Android extracts with the
// executable bit set.
func programPath(dir, name string) string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(dir, name+".exe")
	case "android":
		return filepath.Join(dir, "lib"+name+".so")
	}
	return filepath.Join(dir, name)
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

func withVersion(ctx context.Context, b Binaries) (Binaries, error) {
	b.FFmpeg, b.FFprobe = realProgram(b.FFmpeg), realProgram(b.FFprobe)
	out, err := command(ctx, b.FFmpeg, "-hide_banner", "-version").Output()
	if err != nil {
		return Binaries{}, err
	}
	first, _, _ := strings.Cut(string(out), "\n")
	// "ffmpeg version 7.1-full_build-www.gyan.dev Copyright ..."
	fields := strings.Fields(first)
	if len(fields) >= 3 {
		b.Version = fields[2]
	}
	return b, nil
}
