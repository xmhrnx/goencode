package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"time"
)

// Frame decodes one video frame at a position, scaled to at most maxWidth
// pixels wide. It's used for file thumbnails and the preview comparison.
// inArgs are input options, as for Probe.
func Frame(ctx context.Context, b Binaries, path string, at time.Duration, maxWidth int, inArgs ...string) (image.Image, error) {
	args := []string{"-hide_banner", "-nostdin", "-loglevel", "error"}
	if at > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", at.Seconds()))
	}
	args = append(args, inArgs...)
	args = append(args, "-i", path, "-map", "0:V:0", "-frames:v", "1",
		"-vf", fmt.Sprintf("scale='min(%d,iw)':-2:flags=bicubic,format=rgb24", maxWidth),
		"-f", "image2pipe", "-c:v", "png", "-")
	out, err := command(ctx, b.FFmpeg, args...).Output()
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no frame at %s", FormatDuration(at))
	}
	return png.Decode(bytes.NewReader(out))
}
