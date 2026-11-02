package jobs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xmhrnx/goencode/internal/disc"
)

// MediaExtensions are the file types offered in pickers and collected
// from dropped folders.
var MediaExtensions = []string{
	"mkv", "mp4", "m4v", "mov", "avi", "webm", "wmv", "flv", "ts", "m2ts", "mts", "mpg", "mpeg", "vob", "ogv", "3gp", "mxf", "gif",
	"mp3", "m4a", "aac", "flac", "wav", "ogg", "opus", "wma", "aiff", "aif", "ac3", "eac3", "dts", "mka", "alac",
}

// IsMedia reports whether a path has a media extension.
func IsMedia(path string) bool {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	for _, e := range MediaExtensions {
		if e == ext {
			return true
		}
	}
	return false
}

// Expand replaces folders with the media files inside them, recursively,
// and leaves files as they are. Blu-ray and DVD folders stay folders, one
// per disc, rather than the many clips inside them.
func Expand(paths []string) []string {
	var out []string
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !fi.IsDir() {
			out = append(out, p)
			continue
		}
		if root := disc.Root(p); root != "" {
			out = append(out, root)
			continue
		}
		var found []string
		filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			switch {
			case err != nil:
			case d.IsDir() && path != p && disc.Root(path) == path:
				found = append(found, path)
				return filepath.SkipDir
			case !d.IsDir() && IsMedia(path):
				found = append(found, path)
			}
			return nil
		})
		sort.Strings(found)
		out = append(out, found...)
	}
	return out
}
