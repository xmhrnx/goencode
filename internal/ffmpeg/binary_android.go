//go:build android

package ffmpeg

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const execSupported = true

// platformDirs returns the app's native library folder, where Android puts
// the libffmpeg.so and libffprobe.so packaged in the APK. It's found from the
// path of the app's own Go library in the process's memory map.
func platformDirs() []string {
	f, err := os.Open("/proc/self/maps")
	if err != nil {
		return nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		i := strings.Index(line, "/data/app/")
		if i < 0 || !strings.HasSuffix(line, ".so") {
			continue
		}
		return []string{filepath.Dir(line[i:])}
	}
	return nil
}
