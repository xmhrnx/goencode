package catalog

import "github.com/xmhrnx/goencode/internal/ffmpeg"

func audioBitrate(name string, def, max float64, help string) RateControl {
	return RateControl{ID: "bitrate", Label: "Bitrate", Bitrate: true, Min: 32, Max: max, Step: 8, Default: def,
		Help: name + ": " + help}
}

func lossless(name string) []RateControl {
	return []RateControl{{ID: "lossless", Label: "Lossless", None: true,
		Help: name + " is lossless: the audio is bit-for-bit identical to the source, so there is no bitrate to choose."}}
}

func init() {
	register(Profile{
		Encoder: "aac", Label: "AAC (native)", Kind: ffmpeg.KindAudio,
		Summary: "ffmpeg's own AAC encoder: the standard audio of MP4, M4A and streaming. Good quality from about 128 kb/s for stereo.",
		RateControls: []RateControl{
			audioBitrate("AAC", 160, 512, "128 kb/s is fine for stereo, 192 transparent for most listeners; surround needs about 64 kb/s per channel."),
			{ID: "vbr", Label: "Variable quality", Flag: "q:a", Min: 0.1, Max: 2, Step: 0.1, Default: 1.2,
				Help: "Experimental constant-quality mode of the native AAC encoder; prefer a bitrate."},
		},
		Options: []Option{
			{Key: "aac_coder", Label: "Coder", Kind: OptionChoice,
				Help:    "Two-loop search is the default and best; fast trades quality for speed.",
				Choices: []Choice{{Value: "twoloop", Label: "Two-loop"}, {Value: "fast", Label: "Fast"}}},
			{Key: "profile", Label: "Profile", Kind: OptionChoice,
				Help:    "AAC-LC is what every device plays. MPEG-2 AAC LC is for old broadcast equipment.",
				Choices: []Choice{{Value: "aac_low", Label: "AAC-LC"}, {Value: "mpeg2_aac_low", Label: "MPEG-2 AAC-LC"}}},
		},
		Managed: []string{"q", "b"},
	}, Profile{
		Encoder: "libfdk_aac", Label: "Fraunhofer FDK AAC", Kind: ffmpeg.KindAudio,
		Summary: "The Fraunhofer AAC encoder, the best-sounding AAC encoder, with HE-AAC for low bitrates.",
		RateControls: []RateControl{
			audioBitrate("FDK AAC", 160, 512, "128 kb/s is transparent for most stereo music."),
			{ID: "vbr", Label: "VBR mode", Flag: "vbr", Min: 1, Max: 5, Step: 1, Default: 4,
				Help: "FDK's variable bitrate from 1 (~32 kb/s per channel) to 5 (~112 kb/s per channel)."},
		},
		Options: []Option{
			{Key: "profile", Label: "Profile", Kind: OptionChoice, Simple: true,
				Help:    "HE-AAC sounds better than LC below about 80 kb/s; v2 below about 48 kb/s for stereo.",
				Choices: []Choice{{Value: "aac_low", Label: "AAC-LC"}, {Value: "aac_he", Label: "HE-AAC"}, {Value: "aac_he_v2", Label: "HE-AAC v2"}}},
		},
		Managed: []string{"vbr"},
	}, Profile{
		Encoder: "libopus", Label: "Opus (libopus)", Kind: ffmpeg.KindAudio,
		Summary: "Opus is the best lossy audio codec: transparent music from 128 kb/s, clear speech from 24 kb/s. The audio of WebM and modern web apps.",
		RateControls: []RateControl{
			audioBitrate("Opus", 128, 512, "96 kb/s is good for stereo music, 128 transparent; speech is clear at 24-32 kb/s. Surround: about 48 kb/s per channel."),
		},
		Options: []Option{
			{Key: "vbr", Label: "Bitrate mode", Kind: OptionChoice, Default: "on", Simple: true,
				Help:    "Variable bitrate gives the best quality; constrained VBR keeps the rate steadier for streaming.",
				Choices: []Choice{{Value: "on", Label: "Variable"}, {Value: "constrained", Label: "Constrained"}, {Value: "off", Label: "Constant"}}},
			{Key: "application", Label: "Application", Kind: OptionChoice, Simple: true,
				Help: "Tunes Opus for the content.",
				Choices: []Choice{{Value: "audio", Label: "Music", Help: "Full-band audio, the default."}, {Value: "voip", Label: "Speech", Help: "Improves intelligibility of voice."},
					{Value: "lowdelay", Label: "Low delay"}}},
			{Key: "compression_level", Label: "Complexity", Kind: OptionNumber, Min: 0, Max: 10, Step: 1, Default: "10",
				Help: "Encoder effort from 0 to 10. 10 is the default and only slightly slower."},
		},
	}, Profile{
		Encoder: "opus", Label: "Opus (native)", Kind: ffmpeg.KindAudio,
		Summary:      "ffmpeg's built-in Opus encoder. libopus is better when available.",
		RateControls: []RateControl{audioBitrate("Opus", 128, 512, "96-128 kb/s for stereo music.")},
	}, Profile{
		Encoder: "libmp3lame", Label: "LAME MP3", Kind: ffmpeg.KindAudio,
		Summary: "LAME, the best MP3 encoder. MP3 plays absolutely everywhere.",
		RateControls: []RateControl{
			{ID: "vbr", Label: "Variable quality (VBR)", Flag: "q:a", Min: 0, Max: 9, Step: 1, Default: 2, LowerIsBetter: true,
				Help:  "LAME's V0-V9 scale: V0 averages about 245 kb/s, V2 about 190 kb/s (transparent for most people), V5 about 130 kb/s.",
				Marks: []Mark{{0, "V0"}, {2, "V2"}, {5, "V5"}}},
			audioBitrate("MP3", 192, 320, "constant bitrate. 320 kb/s is the MP3 maximum; VBR gives better quality per byte."),
		},
		Options: []Option{
			{Key: "compression_level", Label: "Algorithm quality", Kind: OptionNumber, Min: 0, Max: 9, Step: 1,
				Help: "LAME's -q: 0 is the slowest and best psychoacoustics, 9 the fastest. The default of 3 is fine."},
			{Key: "joint_stereo", Label: "Joint stereo", Kind: OptionToggle,
				Help: "Codes left/right as mid/side when it saves bits. On by default."},
		},
		Managed: []string{"q", "b"},
	}, Profile{
		Encoder: "libvorbis", Label: "Vorbis", Kind: ffmpeg.KindAudio,
		Summary: "Ogg Vorbis, the open audio codec before Opus.",
		RateControls: []RateControl{
			{ID: "vbr", Label: "Quality", Flag: "q:a", Min: -1, Max: 10, Step: 1, Default: 5,
				Help: "Vorbis quality from -1 to 10; higher is better. 5 averages about 160 kb/s, 3 about 112 kb/s."},
			audioBitrate("Vorbis", 160, 500, "average bitrate."),
		},
		Managed: []string{"q"},
	}, Profile{
		Encoder: "flac", Label: "FLAC", Kind: ffmpeg.KindAudio,
		Summary:      "Free Lossless Audio Codec: perfect quality at about 50-60% of WAV's size.",
		RateControls: lossless("FLAC"),
		Options: []Option{
			{Key: "compression_level", Label: "Compression level", Kind: OptionNumber, Min: 0, Max: 12, Step: 1, Default: "5", Simple: true,
				Help: "0 is fastest, 12 smallest. The audio is identical either way; above 8 gains almost nothing."},
		},
	}, Profile{
		Encoder: "alac", Label: "Apple Lossless", Kind: ffmpeg.KindAudio,
		Summary:      "Apple Lossless (ALAC): lossless audio that plays natively in Apple Music and on every Apple device.",
		RateControls: lossless("ALAC"),
	}, Profile{
		Encoder: "ac3", Label: "Dolby Digital (AC-3)", Kind: ffmpeg.KindAudio,
		Summary:      "Dolby Digital, the surround format of DVDs and broadcast TV. Supported by every AV receiver.",
		RateControls: []RateControl{audioBitrate("AC-3", 448, 640, "448 kb/s for 5.1 is DVD quality, 640 is the maximum; 192 kb/s for stereo.")},
		Options: []Option{
			{Key: "dialnorm", Label: "Dialogue level", Kind: OptionNumber, Min: -31, Max: -1, Step: 1, Default: "-31", Unit: "dB",
				Help: "Average dialogue loudness metadata. Players lower the volume to normalize; -31 means no adjustment."},
		},
	}, Profile{
		Encoder: "eac3", Label: "Dolby Digital Plus (E-AC-3)", Kind: ffmpeg.KindAudio,
		Summary:      "Dolby Digital Plus: more efficient than AC-3, used by streaming services.",
		RateControls: []RateControl{audioBitrate("E-AC-3", 384, 6144, "256-384 kb/s for 5.1, 192 for stereo.")},
	}, Profile{
		Encoder: "mp2", Label: "MPEG Layer II", Kind: ffmpeg.KindAudio,
		Summary:      "MP2, used by DVB broadcast and video CDs.",
		RateControls: []RateControl{audioBitrate("MP2", 224, 384, "192-256 kb/s for stereo.")},
	})

	for _, pcm := range []struct{ enc, label string }{
		{"pcm_s16le", "PCM 16-bit"}, {"pcm_s24le", "PCM 24-bit"}, {"pcm_s32le", "PCM 32-bit"}, {"pcm_f32le", "PCM 32-bit float"},
		{"pcm_s16be", "PCM 16-bit (big-endian)"}, {"pcm_s24be", "PCM 24-bit (big-endian)"},
	} {
		register(Profile{Encoder: pcm.enc, Label: pcm.label, Kind: ffmpeg.KindAudio,
			Summary:      "Uncompressed " + pcm.label + " audio: no quality loss and no decoding cost, but about 10 MB per minute of stereo at 16-bit.",
			RateControls: lossless(pcm.label)})
	}
}
