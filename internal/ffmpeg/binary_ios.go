//go:build ios

package ffmpeg

// iOS doesn't let apps start other programs, so ffmpeg can't run there.
const execSupported = false

func platformDirs() []string { return nil }
