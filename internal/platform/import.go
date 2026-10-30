package platform

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Import copies a picked file's content into the app's import folder and
// returns its path: phones hand apps a stream rather than a path, and
// ffmpeg needs a file.
func Import(name string, r io.Reader) (string, error) {
	if err := os.MkdirAll(ImportDir(), 0o755); err != nil {
		return "", err
	}
	name = filepath.Base(strings.ReplaceAll(name, `\`, "/"))
	if name == "." || name == "/" || name == "" {
		name = "media"
	}
	base := strings.TrimSuffix(name, filepath.Ext(name))
	dst := filepath.Join(ImportDir(), name)
	for i := 2; ; i++ {
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			break
		}
		dst = filepath.Join(ImportDir(), fmt.Sprintf("%s (%d)%s", base, i, filepath.Ext(name)))
	}
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(dst)
		return "", err
	}
	return dst, f.Close()
}
