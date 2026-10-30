//go:build !windows

package platform

// NativePickFiles is only needed on Windows; elsewhere ok is false and the
// app's own picker is used.
func NativePickFiles(title, filterName string, exts []string) (paths []string, ok bool, err error) {
	return nil, false, nil
}
