package settings

import (
	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// Preset is a one-click recipe for the simple mode.
type Preset struct {
	ID          string
	Name        string
	Description string
	Icon        string // icon name in ui/icons
	Tag         string // short badge: "Fast", "Lossless", ...
	// Build returns the recipe for this ffmpeg; output settings are kept
	// from the current recipe by the caller.
	Build func(caps *ffmpeg.Capabilities) Settings
	// Available hides presets this ffmpeg can't do (no GPU encoder).
	Available func(caps *ffmpeg.Capabilities) bool
}

func video(container, codec, encoder, rc string, q float64, opts map[string]string) Settings {
	s := Default()
	s.Container = container
	s.Video.Codec, s.Video.Encoder, s.Video.RateControl, s.Video.Quality = codec, encoder, rc, q
	s.Video.Options = map[string]map[string]string{encoder: opts}
	return s
}

func withAudio(s Settings, codec, encoder string, kbps int) Settings {
	s.Audio.Mode, s.Audio.Codec, s.Audio.Encoder, s.Audio.RateControl, s.Audio.Bitrate = ModeEncode, codec, encoder, "bitrate", kbps
	return s
}

func audioOnly(container, codec, encoder, rc string, q float64, kbps int) Settings {
	s := Default()
	s.Container = container
	s.Video.Mode = ModeDrop
	s.Audio = Audio{Mode: ModeEncode, Tracks: TracksFirst, Codec: codec, Encoder: encoder, RateControl: rc, Quality: q, Bitrate: kbps,
		CopyFallback: true, Options: map[string]map[string]string{}}
	s.Subtitles.Mode = SubsDrop
	return s
}

func hasGPU(caps *ffmpeg.Capabilities) bool {
	for _, c := range []string{"hevc", "h264"} {
		if e := FastestEncoder(caps, c); e != "" {
			if _, hw := ffmpeg.HardwareOf(e); hw {
				return true
			}
		}
	}
	return false
}

// Presets are the recipes of the simple mode, most useful first.
var Presets = []Preset{
	{
		ID: "universal", Name: "Universal MP4", Icon: "device", Tag: "Compatible",
		Description: "H.264 + AAC. Plays on every phone, TV, browser and editor.",
		Build: func(*ffmpeg.Capabilities) Settings {
			return withAudio(video("mp4", "h264", "libx264", "crf", 21, map[string]string{"preset": "medium"}), "aac", "aac", 160)
		},
	},
	{
		ID: "hevc", Name: "Efficient HEVC", Icon: "leaf", Tag: "Smaller",
		Description: "H.265 at half the size of H.264 for the same quality. 10-bit for smooth gradients.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := withAudio(video("mkv", "hevc", "libx265", "crf", 24, map[string]string{"preset": "medium"}), "opus", "libopus", 128)
			s.Video.PixelFormat = "yuv420p10le"
			return s
		},
	},
	{
		ID: "av1", Name: "Modern AV1", Icon: "sparkle", Tag: "Smallest",
		Description: "SVT-AV1 + Opus: the smallest files, for archives and the web.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := withAudio(video("mkv", "av1", "libsvtav1", "crf", 32, map[string]string{"preset": "8"}), "opus", "libopus", 128)
			s.Video.PixelFormat = "yuv420p10le"
			return s
		},
	},
	{
		ID: "gpu", Name: "GPU Turbo", Icon: "zap", Tag: "Fastest",
		Description: "HEVC on your graphics hardware: many times faster than software.",
		Available:   hasGPU,
		Build: func(caps *ffmpeg.Capabilities) Settings {
			enc := FastestEncoder(caps, "hevc")
			codec := "hevc"
			if _, hw := ffmpeg.HardwareOf(enc); !hw {
				codec, enc = "h264", FastestEncoder(caps, "h264")
			}
			s := video("mp4", codec, enc, "", 0, map[string]string{})
			s.Video.HWDecode = true
			return withAudio(s, "aac", "aac", 160)
		},
	},
	{
		ID: "web", Name: "Web WebM", Icon: "globe", Tag: "Open",
		Description: "VP9 + Opus in WebM for HTML5 video, royalty-free.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := withAudio(video("webm", "vp9", "libvpx-vp9", "crf", 32, map[string]string{"deadline": "good", "cpu-used": "2", "row-mt": "1"}), "opus", "libopus", 128)
			return s
		},
	},
	{
		ID: "remux-mkv", Name: "Remux to MKV", Icon: "box", Tag: "Instant",
		Description: "Copy every track and chapter into Matroska. No quality loss.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := Default()
			s.Container = "mkv"
			s.Video.Mode, s.Audio.Mode = ModeCopy, ModeCopy
			return s
		},
	},
	{
		ID: "remux-mp4", Name: "Remux to MP4", Icon: "box", Tag: "Instant",
		Description: "Copy video and audio into MP4; only what MP4 can't hold is converted.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := Default()
			s.Container = "mp4"
			s.Video.Mode, s.Audio.Mode = ModeCopy, ModeCopy
			return s
		},
	},
	{
		ID: "prores", Name: "Editing ProRes", Icon: "scissors", Tag: "Editing",
		Description: "ProRes 422 HQ + PCM in MOV for smooth scrubbing in any editor.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := video("mov", "prores", "prores_ks", "profile", 0, map[string]string{"profile": "hq", "vendor": "apl0"})
			s.Video.PixelFormat = "yuv422p10le"
			s.Audio = Audio{Mode: ModeEncode, Tracks: TracksAll, Codec: "pcm_s16le", Encoder: "pcm_s16le", RateControl: "lossless", CopyFallback: true, Options: map[string]map[string]string{}}
			return s
		},
	},
	{
		ID: "gif", Name: "Animated GIF", Icon: "image", Tag: "Shareable",
		Description: "A looping GIF at 480p and 15 fps with an optimized palette.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := video("gif", "gif", "gif", "palette", 0, map[string]string{})
			s.Video.Height, s.Video.FrameRate = 480, 15
			s.Audio.Mode, s.Subtitles.Mode = ModeDrop, SubsDrop
			return s
		},
	},
	{
		ID: "mp3", Name: "MP3 audio", Icon: "music", Tag: "Audio",
		Description: "Extract the soundtrack as high-quality VBR MP3 (V2).",
		Build: func(*ffmpeg.Capabilities) Settings {
			return audioOnly("mp3", "mp3", "libmp3lame", "vbr", 2, 192)
		},
	},
	{
		ID: "opus", Name: "Opus audio", Icon: "wave", Tag: "Audio",
		Description: "Transparent audio at 128 kb/s, a fraction of MP3's size.",
		Build: func(*ffmpeg.Capabilities) Settings {
			return audioOnly("opus", "opus", "libopus", "bitrate", 0, 128)
		},
	},
	{
		ID: "flac", Name: "FLAC lossless", Icon: "disc", Tag: "Lossless",
		Description: "Bit-perfect audio at about half the size of WAV.",
		Build: func(*ffmpeg.Capabilities) Settings {
			s := audioOnly("flac", "flac", "flac", "lossless", 0, 0)
			s.Audio.Options = map[string]map[string]string{"flac": {"compression_level": "8"}}
			return s
		},
	},
}

// AvailablePresets lists the presets this ffmpeg can run.
func AvailablePresets(caps *ffmpeg.Capabilities) []Preset {
	var out []Preset
	for _, p := range Presets {
		if p.Available == nil || (caps != nil && p.Available(caps)) {
			out = append(out, p)
		}
	}
	return out
}
