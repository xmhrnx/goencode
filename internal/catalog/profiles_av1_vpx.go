package catalog

import "github.com/xmhrnx/goencode/internal/ffmpeg"

func init() {
	register(Profile{
		Encoder: "libsvtav1", Label: "SVT-AV1", Kind: ffmpeg.KindVideo,
		Summary: "SVT-AV1, the fast AV1 encoder from Intel and Netflix: the smallest files of any widely used codec, playable in browsers and recent devices.",
		RateControls: []RateControl{
			{ID: "crf", Label: "Constant quality (CRF)", Flag: "crf", Min: 1, Max: 63, Step: 1, Default: 32, LowerIsBetter: true,
				Help:  "SVT-AV1's quality scale runs from 1 to 63. 35 is its default; 24-30 is high quality, around 40 is small. Not comparable with x264's CRF numbers.",
				Marks: []Mark{{24, "high"}, {35, "default"}, {45, "small"}}},
			{ID: "bitrate", Label: "Target bitrate (VBR)", Bitrate: true, Min: 100, Max: 50000, Step: 100, Default: 2500,
				Help: "SVT-AV1 aims for this average kilobits per second. AV1 needs roughly a third less than HEVC."},
		},
		Options: []Option{
			{Key: "preset", Label: "Preset", Kind: OptionNumber, Min: 0, Max: 13, Step: 1, Default: "10", Simple: true,
				Help: "Speed from 0 (slowest, smallest) to 13 (fastest). 4-6 is good for archiving, 8-10 for everyday encodes; below 4 is very slow."},
			{Key: "svtav1-params", Label: "SVT-AV1 parameters", Kind: OptionText, Simple: true,
				Help: "Raw SVT-AV1 settings, key=value separated by colons. tune=0 favors visual quality, film-grain=8 synthesizes grain instead of encoding it."},
			{Key: "g", Label: "Keyframe interval", Kind: OptionNumber, Min: -1, Max: 1000, Step: 1, Unit: "frames",
				Help: "Distance between keyframes. SVT-AV1 picks about five seconds by default; shorter intervals seek faster."},
		},
		Managed: []string{"crf", "qp", "rc"},
	}, Profile{
		Encoder: "libaom-av1", Label: "libaom AV1", Kind: ffmpeg.KindVideo,
		Summary: "The reference AV1 encoder from the Alliance for Open Media. Excellent compression but very slow; SVT-AV1 is usually the better pick.",
		RateControls: []RateControl{
			{ID: "crf", Label: "Constant quality (CRF)", Flag: "crf", Min: 0, Max: 63, Step: 1, Default: 30, LowerIsBetter: true, Args: []string{"-b:v", "0"},
				Help: "libaom's quality scale from 0 (lossless) to 63. 23-32 is a good range; GoEncode sets the bitrate to 0 for pure constant quality."},
			{ID: "bitrate", Label: "Target bitrate (VBR)", Bitrate: true, Min: 100, Max: 50000, Step: 100, Default: 2500,
				Help: "libaom aims for this average kilobits per second."},
		},
		Options: []Option{
			{Key: "cpu-used", Label: "Speed", Kind: OptionNumber, Min: 0, Max: 8, Step: 1, Default: "4", Simple: true,
				Help: "From 0 (slowest, best) to 8 (fastest). Below 4 libaom is extremely slow; 6 is usable for everyday encodes."},
			{Key: "row-mt", Label: "Row multithreading", Kind: OptionToggle, Simple: true,
				Help: "Encodes rows of tiles in parallel, a big speed-up on multicore computers."},
			{Key: "usage", Label: "Usage", Kind: OptionChoice,
				Help:    "Good quality for files, realtime for streaming.",
				Choices: []Choice{{Value: "good", Label: "Good quality"}, {Value: "realtime", Label: "Realtime"}, {Value: "allintra", Label: "All intra"}}},
		},
		Managed: []string{"crf"},
	}, Profile{
		Encoder: "librav1e", Label: "rav1e", Kind: ffmpeg.KindVideo,
		Summary: "rav1e, a safe and fast AV1 encoder written in Rust.",
		RateControls: []RateControl{
			{ID: "qp", Label: "Quantizer", Flag: "qp", Min: 0, Max: 255, Step: 1, Default: 100, LowerIsBetter: true,
				Help: "rav1e's quantizer from 0 to 255; lower is better. 80-120 is a typical range."},
			{ID: "bitrate", Label: "Target bitrate", Bitrate: true, Min: 100, Max: 50000, Step: 100, Default: 2500,
				Help: "rav1e aims for this average kilobits per second."},
		},
		Options: []Option{
			{Key: "speed", Label: "Speed", Kind: OptionNumber, Min: 0, Max: 10, Step: 1, Default: "6", Simple: true,
				Help: "0 is slowest with the best compression, 10 the fastest."},
		},
		Managed: []string{"qp"},
	}, Profile{
		Encoder: "libvpx-vp9", Label: "libvpx VP9", Kind: ffmpeg.KindVideo,
		Summary: "Google's VP9 encoder, the codec of YouTube and WebM: smaller than H.264 and royalty-free.",
		RateControls: []RateControl{
			{ID: "crf", Label: "Constant quality (CRF)", Flag: "crf", Min: 0, Max: 63, Step: 1, Default: 31, LowerIsBetter: true, Args: []string{"-b:v", "0"},
				Help:  "VP9's quality from 0 to 63. 15-35 is the recommended range; 31 is a good start for 1080p. GoEncode sets the bitrate to 0 so VP9 runs in pure constant quality.",
				Marks: []Mark{{24, "high"}, {31, "1080p"}, {40, "small"}}},
			{ID: "cq", Label: "Constrained quality", Flag: "crf", Min: 0, Max: 63, Step: 1, Default: 31, LowerIsBetter: true, Args: []string{"-b:v", "4M"},
				Help: "Constant quality with a ceiling of 4 Mb/s, so complex scenes can't blow up the size."},
			{ID: "bitrate", Label: "Average bitrate", Bitrate: true, Min: 100, Max: 50000, Step: 100, Default: 3000,
				Help: "libvpx aims for this average kilobits per second."},
		},
		Options: []Option{
			{Key: "deadline", Label: "Deadline", Kind: OptionChoice, Default: "good", Simple: true,
				Help:    "How hard libvpx searches. Good is the right choice for files; best is barely better and much slower.",
				Choices: []Choice{{Value: "good", Label: "Good"}, {Value: "best", Label: "Best"}, {Value: "realtime", Label: "Realtime"}}},
			{Key: "cpu-used", Label: "Speed", Kind: OptionNumber, Min: -8, Max: 8, Step: 1, Default: "1", Simple: true,
				Help: "With the good deadline, 0-5: lower is slower and better. 2-4 is a sensible range for VP9."},
			{Key: "row-mt", Label: "Row multithreading", Kind: OptionToggle, Simple: true,
				Help: "Lets VP9 use many CPU cores on one frame. Strongly recommended."},
			{Key: "tile-columns", Label: "Tile columns", Kind: OptionNumber, Min: 0, Max: 6, Step: 1,
				Help: "Splits frames into 2^n columns encoded in parallel. 2 for 1080p, 3 for 4K."},
			{Key: "lossless", Label: "Lossless", Kind: OptionToggle,
				Help: "Mathematically lossless VP9. Huge files."},
		},
		Managed: []string{"crf"},
	}, Profile{
		Encoder: "libvpx", Label: "libvpx VP8", Kind: ffmpeg.KindVideo,
		Summary: "VP8, the original WebM codec. Use VP9 unless an old browser or device needs VP8.",
		RateControls: []RateControl{
			{ID: "crf", Label: "Constrained quality", Flag: "crf", Min: 4, Max: 63, Step: 1, Default: 10, LowerIsBetter: true, Args: []string{"-b:v", "10M"},
				Help: "VP8's quality from 4 to 63, under a generous bitrate ceiling. 10 is high quality."},
			{ID: "bitrate", Label: "Average bitrate", Bitrate: true, Min: 100, Max: 50000, Step: 100, Default: 3000,
				Help: "libvpx aims for this average kilobits per second."},
		},
		Options: []Option{
			{Key: "deadline", Label: "Deadline", Kind: OptionChoice, Default: "good", Simple: true,
				Help:    "Search effort: good for files, realtime for live.",
				Choices: []Choice{{Value: "good", Label: "Good"}, {Value: "best", Label: "Best"}, {Value: "realtime", Label: "Realtime"}}},
			{Key: "cpu-used", Label: "Speed", Kind: OptionNumber, Min: -16, Max: 16, Step: 1, Simple: true,
				Help: "Higher is faster with lower quality."},
		},
		Managed: []string{"crf"},
	})
}
