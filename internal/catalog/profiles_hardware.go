package catalog

import (
	"strings"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// codecName gives the codec part of a hardware encoder ("hevc_nvenc" ->
// "HEVC"), for help texts.
func codecName(encoder string) string {
	c, _, _ := strings.Cut(encoder, "_")
	switch c {
	case "h264":
		return "H.264"
	case "hevc":
		return "HEVC"
	case "av1":
		return "AV1"
	case "vp9":
		return "VP9"
	case "vp8":
		return "VP8"
	case "mpeg2":
		return "MPEG-2"
	case "mjpeg":
		return "Motion JPEG"
	case "prores":
		return "ProRes"
	}
	return strings.ToUpper(c)
}

func hwBitrate(fam, codec string, extra ...string) RateControl {
	return RateControl{ID: "bitrate", Label: "Average bitrate (VBR)", Bitrate: true, Min: 300, Max: 120000, Step: 100, Default: 6000, Args: extra,
		Help: fam + " aims for this average " + codec + " bitrate in kilobits per second. Hardware encoders need roughly 20-30% more bits than x264/x265 for the same quality."}
}

// hardwareProfile builds the profile of a hardware encoder from its family,
// with help written for that family and codec.
func hardwareProfile(enc string) *Profile {
	fam, ok := ffmpeg.HardwareOf(enc)
	if !ok {
		return nil
	}
	codec := codecName(enc)
	p := &Profile{Encoder: enc, Kind: ffmpeg.KindVideo, Hardware: fam.Name, Curated: true}
	switch fam.Suffix {
	case "_nvenc":
		p.Label = "NVENC " + codec
		p.Summary = "Encodes " + codec + " on the NVIDIA graphics card: many times faster than software with a small cost in efficiency. Ideal for large batches."
		p.RateControls = []RateControl{
			{ID: "cq", Label: "Constant quality (CQ)", Flag: "cq", Min: 1, Max: 51, Step: 1, Default: 24, LowerIsBetter: true, Args: []string{"-rc", "vbr", "-b:v", "0"},
				Help:  "NVENC's constant quality from 1 to 51 in VBR mode; lower is better. 19-23 is high quality for " + codec + ", 28 is small.",
				Marks: []Mark{{19, "high"}, {24, "balanced"}, {30, "small"}}},
			hwBitrate("NVENC", codec, "-rc", "vbr"),
			{ID: "cbr", Label: "Constant bitrate (CBR)", Bitrate: true, Min: 300, Max: 120000, Step: 100, Default: 6000, Args: []string{"-rc", "cbr"},
				Help: "A fixed " + codec + " bitrate for streaming to services that require CBR."},
			{ID: "qp", Label: "Constant quantizer (QP)", Flag: "qp", Min: 0, Max: 51, Step: 1, Default: 23, LowerIsBetter: true, Args: []string{"-rc", "constqp"},
				Help: "The same quantizer for every frame. Simple and predictable quality, less efficient than CQ."},
		}
		p.Options = []Option{
			{Key: "preset", Label: "Preset", Kind: OptionChoice, Default: "p4", Simple: true,
				Help: "NVENC speed preset from P1 (fastest) to P7 (best quality). P5-P7 are still far faster than software encoding.",
				Choices: []Choice{{Value: "p1", Label: "P1 · fastest"}, {Value: "p2", Label: "P2"}, {Value: "p3", Label: "P3"}, {Value: "p4", Label: "P4 · default"},
					{Value: "p5", Label: "P5"}, {Value: "p6", Label: "P6"}, {Value: "p7", Label: "P7 · best quality"}}},
			{Key: "tune", Label: "Tune", Kind: OptionChoice, Simple: true,
				Help: "What NVENC optimizes for.",
				Choices: []Choice{{Value: "hq", Label: "High quality", Help: "For files: the right choice for GoEncode."}, {Value: "uhq", Label: "Ultra high quality", Help: "Slower, newer GPUs only."},
					{Value: "ll", Label: "Low latency"}, {Value: "ull", Label: "Ultra low latency"}, {Value: "lossless", Label: "Lossless"}}},
			{Key: "multipass", Label: "Multipass", Kind: OptionChoice,
				Help:    "Analyzes each frame twice to distribute bits better. Quarter resolution is cheap; full resolution is more accurate.",
				Choices: []Choice{{Value: "disabled", Label: "Off"}, {Value: "qres", Label: "Quarter resolution"}, {Value: "fullres", Label: "Full resolution"}}},
			{Key: "spatial-aq", Label: "Spatial AQ", Kind: OptionToggle,
				Help: "Moves bits from busy to flat areas within a frame, reducing banding and blocking in skies and walls."},
			{Key: "temporal-aq", Label: "Temporal AQ", Kind: OptionToggle,
				Help: "Moves bits to parts of the picture that stay still across frames. Helps slow scenes."},
			{Key: "rc-lookahead", Label: "Lookahead", Kind: OptionNumber, Min: 0, Max: 64, Step: 1, Unit: "frames",
				Help: "Frames NVENC looks ahead to plan bits and place B-frames. 20-32 improves quality at a small speed cost."},
			{Key: "bf", Label: "B-frames", Kind: OptionNumber, Min: 0, Max: 4, Step: 1,
				Help: "Bidirectional frames between references. Turing and newer GPUs support them for " + codec + "."},
		}
		p.Managed = []string{"cq", "qp", "rc"}
	case "_qsv":
		p.Label = "Quick Sync " + codec
		p.Summary = "Encodes " + codec + " on the Intel GPU (Quick Sync Video): fast and power-efficient, good quality on recent Intel chips."
		p.RateControls = []RateControl{
			{ID: "icq", Label: "Intelligent constant quality", Flag: "global_quality", Min: 1, Max: 51, Step: 1, Default: 24, LowerIsBetter: true,
				Help: "Quick Sync's ICQ mode: constant visual quality from 1 to 51, lower is better. 20-25 is a good range for " + codec + "."},
			hwBitrate("Quick Sync", codec),
		}
		p.Options = []Option{
			{Key: "preset", Label: "Preset", Kind: OptionChoice, Default: "medium", Simple: true,
				Help: "Quick Sync speed/quality trade-off.",
				Choices: []Choice{{Value: "veryfast", Label: "Very fast"}, {Value: "faster", Label: "Faster"}, {Value: "fast", Label: "Fast"},
					{Value: "medium", Label: "Medium"}, {Value: "slow", Label: "Slow"}, {Value: "slower", Label: "Slower"}, {Value: "veryslow", Label: "Very slow"}}},
			{Key: "look_ahead_depth", Label: "Lookahead", Kind: OptionNumber, Min: 0, Max: 100, Step: 1, Unit: "frames",
				Help: "Frames analyzed ahead to distribute bits. Improves quality in scenes with motion."},
		}
		p.Managed = []string{"global_quality"}
	case "_amf":
		p.Label = "AMF " + codec
		p.Summary = "Encodes " + codec + " on the AMD graphics card through AMF: fast hardware encoding."
		p.RateControls = []RateControl{
			{ID: "cqp", Label: "Constant quantizer (CQP)", Flag: "qp_i", AlsoFlags: []string{"qp_p", "qp_b"}, Min: 0, Max: 51, Step: 1, Default: 22, LowerIsBetter: true, Args: []string{"-rc", "cqp"},
				Help: "The same quantizer for I, P and B frames; lower is better. 18-24 is high quality for " + codec + "."},
			hwBitrate("AMF", codec, "-rc", "vbr_peak"),
			{ID: "cbr", Label: "Constant bitrate (CBR)", Bitrate: true, Min: 300, Max: 120000, Step: 100, Default: 6000, Args: []string{"-rc", "cbr"},
				Help: "A fixed bitrate for live streaming."},
		}
		p.Options = []Option{
			{Key: "quality", Label: "Quality preset", Kind: OptionChoice, Default: "speed", Simple: true,
				Help:    "AMF's speed/quality trade-off. Quality is noticeably better and still fast.",
				Choices: []Choice{{Value: "speed", Label: "Speed"}, {Value: "balanced", Label: "Balanced"}, {Value: "quality", Label: "Quality"}}},
			{Key: "preanalysis", Label: "Pre-analysis", Kind: OptionToggle,
				Help: "Analyzes frames before encoding for better bit distribution."},
		}
		p.Managed = []string{"rc", "qp_i", "qp_p", "qp_b"}
	case "_videotoolbox":
		p.Label = "VideoToolbox " + codec
		p.Summary = "Apple's hardware " + codec + " encoder on the Mac's media engine: very fast and power-efficient."
		p.RateControls = []RateControl{
			{ID: "quality", Label: "Constant quality", Flag: "q:v", Min: 1, Max: 100, Step: 1, Default: 60,
				Help:  "VideoToolbox quality from 1 to 100; higher is better (the opposite of CRF). 55-65 suits most video. Requires Apple Silicon.",
				Marks: []Mark{{50, "small"}, {65, "high"}, {80, "very high"}}},
			hwBitrate("VideoToolbox", codec),
		}
		p.Options = []Option{
			{Key: "allow_sw", Label: "Allow software fallback", Kind: OptionToggle, Simple: true,
				Help: "Lets macOS fall back to its software encoder when the hardware is busy or doesn't support the settings."},
			{Key: "realtime", Label: "Realtime", Kind: OptionToggle,
				Help: "Prioritizes speed for live capture over quality."},
			{Key: "prio_speed", Label: "Prioritize speed", Kind: OptionToggle,
				Help: "Faster encoding at slightly lower quality."},
		}
	case "_vaapi":
		p.Label = "VA-API " + codec
		p.Summary = "Encodes " + codec + " on the GPU through VA-API (Intel and AMD on Linux)."
		p.RateControls = []RateControl{
			{ID: "cqp", Label: "Constant quantizer (CQP)", Flag: "qp", Min: 0, Max: 52, Step: 1, Default: 24, LowerIsBetter: true, Args: []string{"-rc_mode", "CQP"},
				Help: "A fixed quantizer for every frame; lower is better. 20-26 suits " + codec + "."},
			{ID: "icq", Label: "Intelligent constant quality", Flag: "global_quality", Min: 1, Max: 51, Step: 1, Default: 25, LowerIsBetter: true, Args: []string{"-rc_mode", "ICQ"},
				Help: "Constant visual quality where the driver supports ICQ (Intel)."},
			hwBitrate("VA-API", codec, "-rc_mode", "VBR"),
		}
		p.Options = []Option{
			{Key: "compression_level", Label: "Compression level", Kind: OptionNumber, Min: 0, Max: 7, Step: 1, Simple: true,
				Help: "Driver-specific speed/quality trade-off; lower is usually slower and better."},
		}
		p.Managed = []string{"rc_mode", "qp", "global_quality"}
	case "_mediacodec":
		p.Label = "MediaCodec " + codec
		p.Summary = "Encodes " + codec + " with the phone's hardware video encoder: fast and battery-friendly."
		p.RateControls = []RateControl{
			hwBitrate("MediaCodec", codec),
			{ID: "cq", Label: "Constant quality", Bitrate: true, Min: 300, Max: 80000, Step: 100, Default: 8000, Args: []string{"-bitrate_mode", "cq"},
				Help: "Asks the encoder for constant quality where the device supports it, with this bitrate as a guide."},
		}
		p.Managed = []string{"bitrate_mode"}
	case "_mf":
		p.Label = "Media Foundation " + codec
		p.Summary = "Encodes " + codec + " through Windows Media Foundation, using whatever hardware encoder Windows provides."
		p.RateControls = []RateControl{
			{ID: "quality", Label: "Quality", Flag: "quality", Min: 0, Max: 100, Step: 1, Default: 70, Args: []string{"-rate_control", "quality"},
				Help: "Media Foundation's quality from 0 to 100; higher is better."},
			hwBitrate("Media Foundation", codec, "-rate_control", "u_vbr"),
		}
		p.Options = []Option{
			{Key: "hw_encoding", Label: "Force hardware", Kind: OptionToggle, Simple: true,
				Help: "Only use a hardware encoder; fail instead of falling back to Microsoft's software encoder."},
		}
		p.Managed = []string{"rate_control", "quality"}
	default:
		return nil
	}
	return p
}
