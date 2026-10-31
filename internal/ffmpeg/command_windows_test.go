//go:build windows

package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRealProgram(t *testing.T) {
	touch := func(p string) string {
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, nil, 0o644)
		return p
	}
	root := filepath.Join(t.TempDir(), "chocolatey")
	shim := touch(filepath.Join(root, "bin", "ffmpeg.exe"))
	real := touch(filepath.Join(root, "lib", "ffmpeg", "tools", "ffmpeg", "bin", "ffmpeg.exe"))
	if got := realProgram(shim); got != real {
		t.Errorf("chocolatey: got %s, want %s", got, real)
	}

	scoop := filepath.Join(t.TempDir(), "scoop")
	sshim := touch(filepath.Join(scoop, "shims", "ffmpeg.exe"))
	sreal := touch(filepath.Join(scoop, "apps", "ffmpeg", "current", "bin", "ffmpeg.exe"))
	os.WriteFile(filepath.Join(scoop, "shims", "ffmpeg.shim"), []byte(`path = "`+sreal+`"`+"\r\n"), 0o644)
	if got := realProgram(sshim); got != sreal {
		t.Errorf("scoop: got %s, want %s", got, sreal)
	}

	plain := touch(filepath.Join(t.TempDir(), "tools", "ffmpeg.exe"))
	if got := realProgram(plain); got != plain {
		t.Errorf("plain: got %s", got)
	}
}
