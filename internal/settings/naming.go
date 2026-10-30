package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmhrnx/goencode/internal/catalog"
)

// OutputPath works out where a file's result goes. fallbackDir is used
// when the recipe says "next to the source" but the source has no folder
// the app may write to (files imported on a phone). taken reports paths
// already claimed by other jobs in the batch.
func OutputPath(input string, s Settings, fallbackDir string, taken func(string) bool) string {
	c := catalog.ContainerByID(s.Container)
	dir := s.Output.Dir
	if dir == "" {
		dir = filepath.Dir(input)
		if fallbackDir != "" {
			dir = fallbackDir
		}
	}
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	name := base + s.Output.Suffix
	out := filepath.Join(dir, name+"."+c.Ext)
	if samePath(out, input) {
		name += "-1"
		out = filepath.Join(dir, name+"."+c.Ext)
	}
	if s.Output.Overwrite && !samePath(out, input) && (taken == nil || !taken(out)) {
		return out
	}
	for i := 2; exists(out) || (taken != nil && taken(out)); i++ {
		out = filepath.Join(dir, fmt.Sprintf("%s (%d).%s", name, i, c.Ext))
	}
	return out
}

func samePath(a, b string) bool {
	pa, _ := filepath.Abs(a)
	pb, _ := filepath.Abs(b)
	return strings.EqualFold(pa, pb)
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
