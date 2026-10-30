package catalog

import "github.com/xmhrnx/goencode/internal/ffmpeg"

func init() {
	register(Profile{
		Encoder: "prores_ks", Label: "ProRes (Kostya)", Kind: ffmpeg.KindVideo,
		Summary: "Apple ProRes, the editing codec: fast to cut and scrub, very high quality, large files. Best in MOV.",
		RateControls: []RateControl{{ID: "profile", Label: "Set by profile", None: true,
			Help: "ProRes has no quality slider: its bitrate is fixed by the profile below."}},
		Options: []Option{
			{Key: "profile", Label: "Profile", Kind: OptionChoice, Default: "hq", Simple: true,
				Help: "ProRes flavor. 422 HQ is the usual mastering choice; 4444 keeps full color and an alpha channel.",
				Choices: []Choice{
					{Value: "proxy", Label: "422 Proxy", Help: "About 45 Mb/s at 1080p: offline editing proxies."},
					{Value: "lt", Label: "422 LT", Help: "About 100 Mb/s at 1080p."},
					{Value: "standard", Label: "422", Help: "About 145 Mb/s at 1080p."},
					{Value: "hq", Label: "422 HQ", Help: "About 220 Mb/s at 1080p: visually lossless mastering."},
					{Value: "4444", Label: "4444", Help: "Full chroma resolution with alpha."},
					{Value: "4444xq", Label: "4444 XQ", Help: "The highest-quality ProRes."},
				}},
			{Key: "vendor", Label: "Vendor tag", Kind: OptionChoice,
				Help:    "Writes Apple's vendor ID so picky Apple software accepts the file as genuine ProRes.",
				Choices: []Choice{{Value: "apl0", Label: "Apple (apl0)"}}},
		},
	}, Profile{
		Encoder: "dnxhd", Label: "DNxHD / DNxHR", Kind: ffmpeg.KindVideo,
		Summary: "Avid's editing codec, DNxHR for any resolution. Native to Avid Media Composer and well supported by Resolve and Premiere.",
		RateControls: []RateControl{{ID: "profile", Label: "Set by profile", None: true,
			Help: "DNxHR's bitrate follows the profile; GoEncode uses the DNxHR family, which works at any resolution."}},
		Options: []Option{
			{Key: "profile", Label: "Profile", Kind: OptionChoice, Default: "dnxhr_hq", Simple: true,
				Help: "DNxHR quality class.",
				Choices: []Choice{
					{Value: "dnxhr_lb", Label: "LB · low bandwidth", Help: "Offline quality proxies."},
					{Value: "dnxhr_sq", Label: "SQ · standard"},
					{Value: "dnxhr_hq", Label: "HQ · high quality", Help: "8-bit mastering."},
					{Value: "dnxhr_hqx", Label: "HQX · 10-bit", Help: "10-bit 4:2:2 for grading."},
					{Value: "dnxhr_444", Label: "444", Help: "10-bit full chroma."},
				}},
		},
	}, Profile{
		Encoder: "ffv1", Label: "FFV1", Kind: ffmpeg.KindVideo,
		Summary:      "FFV1 is a mathematically lossless codec trusted by archives and libraries. Very large files, perfect quality.",
		RateControls: []RateControl{{ID: "lossless", Label: "Lossless", None: true, Help: "FFV1 is always lossless; there is no quality setting."}},
		Options: []Option{
			{Key: "level", Label: "Version", Kind: OptionChoice, Default: "3", Simple: true,
				Help:    "FFV1 version 3 adds multithreading and checksums for archival integrity.",
				Choices: []Choice{{Value: "1", Label: "Version 1"}, {Value: "3", Label: "Version 3"}}},
			{Key: "slicecrc", Label: "Slice checksums", Kind: OptionToggle, Simple: true,
				Help: "Stores a checksum per slice so damage can be detected later."},
			{Key: "coder", Label: "Coder", Kind: OptionChoice,
				Help:    "Range coder compresses better than Golomb-Rice.",
				Choices: []Choice{{Value: "rice", Label: "Golomb-Rice"}, {Value: "range_def", Label: "Range coder"}}},
		},
	}, Profile{
		Encoder: "mpeg4", Label: "MPEG-4 Part 2", Kind: ffmpeg.KindVideo,
		Summary: "MPEG-4 Part 2 (DivX/Xvid style): for old DVD players and devices that predate H.264.",
		RateControls: []RateControl{
			{ID: "qscale", Label: "Quantizer scale", Flag: "q:v", Min: 1, Max: 31, Step: 1, Default: 4, LowerIsBetter: true,
				Help: "MPEG-4 quality from 1 (best) to 31 (worst). 3-5 is good quality."},
			{ID: "bitrate", Label: "Average bitrate", Bitrate: true, Min: 200, Max: 20000, Step: 100, Default: 2500,
				Help: "Average kilobits per second."},
		},
		Options: []Option{
			{Key: "vtag", Label: "FourCC tag", Kind: OptionChoice, Simple: true,
				Help:    "Labels the stream as Xvid or DivX so old hardware players recognize it.",
				Choices: []Choice{{Value: "xvid", Label: "Xvid"}, {Value: "DIVX", Label: "DivX"}}},
		},
		Managed: []string{"q", "qscale"},
	}, Profile{
		Encoder: "mpeg2video", Label: "MPEG-2", Kind: ffmpeg.KindVideo,
		Summary: "MPEG-2 video, the DVD and broadcast codec.",
		RateControls: []RateControl{
			{ID: "qscale", Label: "Quantizer scale", Flag: "q:v", Min: 1, Max: 31, Step: 1, Default: 3, LowerIsBetter: true,
				Help: "MPEG-2 quality from 1 (best) to 31 (worst)."},
			{ID: "bitrate", Label: "Average bitrate", Bitrate: true, Min: 500, Max: 80000, Step: 100, Default: 8000,
				Help: "DVDs use up to about 9800 kb/s."},
		},
		Managed: []string{"q", "qscale"},
	}, Profile{
		Encoder: "mjpeg", Label: "Motion JPEG", Kind: ffmpeg.KindVideo,
		Summary: "Every frame is a JPEG: simple to edit and decode, big files.",
		RateControls: []RateControl{{ID: "qscale", Label: "JPEG quality", Flag: "q:v", Min: 2, Max: 31, Step: 1, Default: 3, LowerIsBetter: true,
			Help: "2 is the best JPEG quality, 31 the worst."}},
		Managed: []string{"q", "qscale"},
	}, Profile{
		Encoder: "gif", Label: "GIF", Kind: ffmpeg.KindVideo,
		Summary: "Animated GIF. GoEncode builds an optimized 256-color palette for each file.",
		RateControls: []RateControl{{ID: "palette", Label: "Optimized palette", None: true,
			Help: "GIF has no quality setting; colors come from a palette GoEncode generates from the video. Lower the frame rate and size for smaller files."}},
	}, Profile{
		Encoder: "libtheora", Label: "Theora", Kind: ffmpeg.KindVideo,
		Summary: "Theora, the original Ogg video codec.",
		RateControls: []RateControl{
			{ID: "qscale", Label: "Quality", Flag: "q:v", Min: 0, Max: 10, Step: 1, Default: 7,
				Help: "Theora quality from 0 to 10; higher is better. 6-8 is good."},
			{ID: "bitrate", Label: "Average bitrate", Bitrate: true, Min: 200, Max: 20000, Step: 100, Default: 3000, Help: "Average kilobits per second."},
		},
		Managed: []string{"q", "qscale"},
	})
}
