package catalog

import "github.com/xmhrnx/goencode/internal/ffmpeg"

func x26xPresets(name string) []Choice {
	return []Choice{
		{Value: "ultrafast", Label: "Ultrafast", Help: "Fastest " + name + " preset; files are much bigger for the same quality."},
		{Value: "superfast", Label: "Superfast"},
		{Value: "veryfast", Label: "Very fast", Help: "A good choice for quick encodes and streaming."},
		{Value: "faster", Label: "Faster"},
		{Value: "fast", Label: "Fast"},
		{Value: "medium", Label: "Medium", Help: "The " + name + " default: balanced speed and compression."},
		{Value: "slow", Label: "Slow", Help: "About 40% slower than medium for files around 5-10% smaller."},
		{Value: "slower", Label: "Slower"},
		{Value: "veryslow", Label: "Very slow", Help: "Near the best compression " + name + " offers; for archiving when time doesn't matter."},
		{Value: "placebo", Label: "Placebo", Help: "Extremely slow for a gain you can't see. Not recommended."},
	}
}

func init() {
	register(Profile{
		Encoder: "libx264", Label: "x264", Kind: ffmpeg.KindVideo,
		Summary: "x264, the reference H.264 encoder: excellent quality and plays on every device made in the last fifteen years.",
		RateControls: []RateControl{
			{ID: "crf", Label: "Constant quality (CRF)", Flag: "crf", Min: 0, Max: 51, Step: 1, Default: 21, LowerIsBetter: true,
				Help:  "x264 keeps the perceived quality constant and spends bits where the picture needs them. 0 is lossless, 18 looks visually lossless, 23 is the default; each +6 roughly halves the file size.",
				Marks: []Mark{{18, "visually lossless"}, {23, "default"}, {28, "small"}}},
			{ID: "bitrate", Label: "Average bitrate (ABR)", Bitrate: true, Min: 200, Max: 80000, Step: 100, Default: 6000,
				Help: "x264 aims for this average kilobits per second. Predictable size, but quality varies with the content; prefer CRF unless a size limit matters."},
			{ID: "qp", Label: "Constant quantizer (QP)", Flag: "qp", Min: 0, Max: 69, Step: 1, Default: 23, LowerIsBetter: true,
				Help: "Every frame uses the same quantizer. Mainly for testing; CRF gives better quality at the same size. 0 is lossless."},
		},
		Options: []Option{
			{Key: "preset", Label: "Preset", Kind: OptionChoice, Default: "medium", Simple: true, Choices: x26xPresets("x264"),
				Help: "Trades encoding speed for compression. Slower presets make smaller files at the same CRF; quality stays the same."},
			{Key: "tune", Label: "Tune", Kind: OptionChoice, Simple: true,
				Help: "Adjusts x264's psychovisual settings for a kind of content.",
				Choices: []Choice{
					{Value: "film", Label: "Film", Help: "Live-action footage: keeps fine detail and grain a little better."},
					{Value: "animation", Label: "Animation", Help: "Cartoons and anime: more reference frames and deblocking for flat areas."},
					{Value: "grain", Label: "Grain", Help: "Keeps heavy film grain instead of smoothing it away. Bigger files."},
					{Value: "stillimage", Label: "Still image", Help: "Slideshows and mostly static content."},
					{Value: "fastdecode", Label: "Fast decode", Help: "Disables CABAC and other features so weak devices can play it."},
					{Value: "zerolatency", Label: "Zero latency", Help: "No lookahead or B-frames, for live streaming and video calls."},
				}},
			{Key: "profile", Label: "Profile", Kind: OptionChoice,
				Help: "Restricts H.264 features for compatibility with older devices. High is supported by everything recent.",
				Choices: []Choice{
					{Value: "baseline", Label: "Baseline", Help: "Very old phones and video conferencing. No B-frames or CABAC."},
					{Value: "main", Label: "Main", Help: "Older set-top boxes and DVD-era players."},
					{Value: "high", Label: "High", Help: "Every modern device: the default for 8-bit video."},
					{Value: "high10", Label: "High 10", Help: "10-bit color. Fewer hardware decoders support it."},
					{Value: "high422", Label: "High 4:2:2", Help: "4:2:2 chroma, for production workflows."},
					{Value: "high444", Label: "High 4:4:4", Help: "Full chroma and lossless support. Rarely hardware-decodable."},
				}},
			{Key: "level", Label: "Level", Kind: OptionChoice,
				Help:    "Caps resolution, frame rate and bitrate so a device's decoder can keep up. Leave on auto unless a device demands a level.",
				Choices: levels("3.0", "3.1", "4.0", "4.1", "4.2", "5.0", "5.1", "5.2")},
			{Key: "x264-params", Label: "x264 parameters", Kind: OptionText,
				Help: "Raw x264 settings as key=value pairs separated by colons, e.g. aq-mode=3:psy-rd=1.0,0.15. Applied after the preset and tune."},
			{Key: "g", Label: "Keyframe interval", Kind: OptionNumber, Min: 1, Max: 1000, Step: 1, Unit: "frames",
				Help: "Maximum distance between keyframes. Smaller values make seeking faster and files bigger; x264 defaults to 250."},
			{Key: "bf", Label: "B-frames", Kind: OptionNumber, Min: 0, Max: 16, Step: 1,
				Help: "Maximum consecutive bidirectional frames. More B-frames compress better; the preset already picks a sensible number."},
			{Key: "refs", Label: "Reference frames", Kind: OptionNumber, Min: 1, Max: 16, Step: 1,
				Help: "How many past frames each frame may reference. More helps compression and slows encoding; some devices limit it."},
			{Key: "aq-mode", Label: "Adaptive quantization", Kind: OptionChoice,
				Help: "How x264 redistributes bits between flat and detailed areas of a frame.",
				Choices: []Choice{
					{Value: "0", Label: "Off"},
					{Value: "1", Label: "Variance", Help: "x264's default."},
					{Value: "2", Label: "Auto-variance"},
					{Value: "3", Label: "Auto-variance, dark bias", Help: "Spends more bits in dark scenes, reducing banding in shadows."},
				}},
		},
		Managed: []string{"crf", "qp", "crf_max", "x264opts", "passlogfile", "fastfirstpass"},
	})

	register(Profile{
		Encoder: "libx265", Label: "x265", Kind: ffmpeg.KindVideo,
		Summary: "x265 encodes HEVC (H.265): about half the size of H.264 at the same quality, at a much slower speed. Plays on modern TVs, phones and Apple devices.",
		RateControls: []RateControl{
			{ID: "crf", Label: "Constant quality (CRF)", Flag: "crf", Min: 0, Max: 51, Step: 1, Default: 24, LowerIsBetter: true,
				Help:  "x265 keeps the perceived quality constant. Its scale differs from x264's: 28 is the default and looks roughly like x264 at 23; 20-24 is high quality.",
				Marks: []Mark{{20, "high"}, {28, "default"}, {32, "small"}}},
			{ID: "bitrate", Label: "Average bitrate (ABR)", Bitrate: true, Min: 200, Max: 80000, Step: 100, Default: 4000,
				Help: "x265 aims for this average kilobits per second. HEVC needs about half the bitrate of H.264 for the same quality."},
			{ID: "qp", Label: "Constant quantizer (QP)", Flag: "qp", Min: 0, Max: 51, Step: 1, Default: 28, LowerIsBetter: true,
				Help: "A fixed quantizer for every frame. For testing; CRF is better in practice."},
		},
		Options: []Option{
			{Key: "preset", Label: "Preset", Kind: OptionChoice, Default: "medium", Simple: true, Choices: x26xPresets("x265"),
				Help: "Trades encoding speed for compression. x265 is slow: fast or medium is reasonable for most uses, slow for archives."},
			{Key: "tune", Label: "Tune", Kind: OptionChoice, Simple: true,
				Help: "Adjusts x265 for a kind of content or goal.",
				Choices: []Choice{
					{Value: "grain", Label: "Grain", Help: "Preserves film grain instead of smearing it."},
					{Value: "animation", Label: "Animation", Help: "Cartoons and anime."},
					{Value: "fastdecode", Label: "Fast decode", Help: "Easier to play on weak hardware."},
					{Value: "zerolatency", Label: "Zero latency", Help: "No frame delay, for live use."},
					{Value: "psnr", Label: "PSNR", Help: "Optimizes for the PSNR metric rather than for the eye."},
					{Value: "ssim", Label: "SSIM", Help: "Optimizes for the SSIM metric rather than for the eye."},
				}},
			{Key: "profile", Label: "Profile", Kind: OptionChoice,
				Help: "HEVC profile. Main is 8-bit, Main 10 is 10-bit, which reduces banding and is required for HDR.",
				Choices: []Choice{
					{Value: "main", Label: "Main", Help: "8-bit 4:2:0, the most compatible."},
					{Value: "main10", Label: "Main 10", Help: "10-bit 4:2:0: smoother gradients, needed for HDR10. Supported by most HEVC devices."},
					{Value: "main12", Label: "Main 12", Help: "12-bit, rarely decodable in hardware."},
					{Value: "main422-10", Label: "Main 4:2:2 10"},
					{Value: "main444-8", Label: "Main 4:4:4"},
				}},
			{Key: "x265-params", Label: "x265 parameters", Kind: OptionText,
				Help: "Raw x265 settings as key=value pairs separated by colons, e.g. aq-mode=3:no-sao=1. Applied after the preset and tune."},
			{Key: "g", Label: "Keyframe interval", Kind: OptionNumber, Min: 1, Max: 1000, Step: 1, Unit: "frames",
				Help: "Maximum distance between keyframes. x265 defaults to 250; streaming often uses two seconds' worth."},
			{Key: "bf", Label: "B-frames", Kind: OptionNumber, Min: 0, Max: 16, Step: 1,
				Help: "Maximum consecutive B-frames. The preset already chooses well."},
		},
		Managed: []string{"crf", "qp", "forced-idr", "udu_sei", "a53cc", "dolbyvision"},
	})
}

func levels(vals ...string) []Choice {
	cs := make([]Choice, len(vals))
	for i, v := range vals {
		cs[i] = Choice{Value: v, Label: v}
	}
	return cs
}
