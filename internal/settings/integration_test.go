package settings

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// TestRealEncodes runs real recipes through the installed ffmpeg on a
// generated file with two audio tracks, subtitles and chapters. It's
// skipped when ffmpeg isn't installed.
func TestRealEncodes(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	ctx := context.Background()
	bins, err := ffmpeg.Locate(ctx, "")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	caps, err := ffmpeg.Detect(ctx, bins)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	src := makeSample(t, bins, dir)

	info, err := ffmpeg.Probe(ctx, bins, src)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.StreamsOf(ffmpeg.KindAudio)) != 2 || len(info.StreamsOf(ffmpeg.KindSubtitle)) != 1 || info.Chapters != 2 {
		t.Fatalf("sample probed wrong: %+v", info)
	}

	cases := map[string]func() Settings{
		"remux-mp4": func() Settings { return presetByID("remux-mp4").Build(caps) },
		"remux-mkv": func() Settings { return presetByID("remux-mkv").Build(caps) },
		"universal": func() Settings { return presetByID("universal").Build(caps) },
		"mp3":       func() Settings { return presetByID("mp3").Build(caps) },
		"flac":      func() Settings { return presetByID("flac").Build(caps) },
		"gif":       func() Settings { return presetByID("gif").Build(caps) },
		"webm":      func() Settings { return presetByID("web").Build(caps) },
		"burn": func() Settings {
			s := Default()
			s.Subtitles.Mode = SubsBurn
			s.Video.Height = 120
			return s
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			s := Normalize(build(), caps)
			out := filepath.Join(dir, name+"."+extOf(s))
			args, plan, dur, err := Args(Job{Input: src, Info: info, Output: out, Window: Window{Start: time.Second, Duration: 2 * time.Second}}, s, caps)
			if err != nil {
				t.Fatal(err)
			}
			var last ffmpeg.Progress
			if err := ffmpeg.Run(ctx, bins, args, dur, func(p ffmpeg.Progress) { last = p }); err != nil {
				t.Fatalf("%v\nargs: %q", err, args)
			}
			if last.Fraction != 1 {
				t.Errorf("final progress = %v", last.Fraction)
			}
			got, err := ffmpeg.Probe(ctx, bins, out)
			if err != nil {
				t.Fatal(err)
			}
			if plan.Video != nil {
				if _, ok := got.Video(); !ok {
					t.Error("output has no video")
				}
			}
			if n := len(got.StreamsOf(ffmpeg.KindAudio)); n != len(nonDropped(plan.Audio)) {
				t.Errorf("output has %d audio tracks, plan said %d", n, len(nonDropped(plan.Audio)))
			}
		})
	}
}

func presetByID(id string) Preset {
	for _, p := range Presets {
		if p.ID == id {
			return p
		}
	}
	panic(id)
}

func extOf(s Settings) string {
	return map[string]string{"mp4": "mp4", "mkv": "mkv", "mp3": "mp3", "flac": "flac", "gif": "gif", "webm": "webm"}[s.Container]
}

func makeSample(t *testing.T, b ffmpeg.Binaries, dir string) string {
	t.Helper()
	srt := filepath.Join(dir, "subs.srt")
	os.WriteFile(srt, []byte("1\n00:00:00,500 --> 00:00:03,500\nHello GoEncode\n\n2\n00:00:03,600 --> 00:00:05,000\nSecond line\n"), 0o644)
	meta := filepath.Join(dir, "meta.txt")
	os.WriteFile(meta, []byte(";FFMETADATA1\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=2500\ntitle=One\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=2500\nEND=5000\ntitle=Two\n"), 0o644)
	out := filepath.Join(dir, "sample.mkv")
	cmd := exec.Command(b.FFmpeg, "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc2=s=320x240:r=25:d=5",
		"-f", "lavfi", "-i", "sine=f=440:d=5",
		"-f", "lavfi", "-i", "sine=f=660:d=5",
		"-i", srt, "-i", meta,
		"-map", "0", "-map", "1", "-map", "2", "-map", "3", "-map_chapters", "4",
		"-c:v", "libx264", "-preset", "ultrafast", "-c:a", "aac", "-c:s", "srt",
		"-metadata:s:a:0", "language=eng", "-metadata:s:a:1", "language=jpn",
		"-y", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("making sample: %v\n%s", err, b)
	}
	return out
}
