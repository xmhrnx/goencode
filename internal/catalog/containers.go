package catalog

// Containers are the output formats GoEncode offers, video first.
var Containers = []Container{
	{
		ID: "mkv", Label: "MKV", Ext: "mkv", Muxer: "matroska",
		Description: "Matroska holds almost any codec, any number of audio and subtitle tracks and chapters. The safest choice for remuxing.",
		Video:       []string{"*"}, Audio: []string{"*"},
		Subtitle:    []string{"subrip", "ass", "ssa", "webvtt", "text", "dvd_subtitle", "hdmv_pgs_subtitle", "dvb_subtitle"},
		Chapters:    true,
		PreferVideo: []string{"hevc", "h264", "av1"}, PreferAudio: []string{"opus", "aac", "flac"}, PreferSubtitle: []string{"ass", "subrip"},
	},
	{
		ID: "mp4", Label: "MP4", Ext: "mp4", Muxer: "mp4",
		Description: "MPEG-4 plays on virtually every phone, browser and TV. Subtitles become mov_text; picture subtitles can't be stored.",
		Video:       []string{"h264", "hevc", "av1", "vp9", "mpeg4", "mpeg2video", "mpeg1video", "mjpeg", "png", "vvc"},
		Audio:       []string{"aac", "mp3", "ac3", "eac3", "opus", "flac", "alac", "mp2", "truehd", "dts"},
		Subtitle:    []string{"mov_text"},
		Chapters:    true,
		PreferVideo: []string{"h264", "hevc", "av1"}, PreferAudio: []string{"aac", "ac3", "opus"}, PreferSubtitle: []string{"mov_text"},
		Args: []string{"-movflags", "+faststart"},
	},
	{
		ID: "mov", Label: "MOV", Ext: "mov", Muxer: "mov",
		Description: "QuickTime, the native format of Final Cut and DaVinci Resolve: ProRes, DNxHD and uncompressed audio for editing.",
		Video:       []string{"h264", "hevc", "prores", "dnxhd", "mjpeg", "mpeg4", "png", "qtrle", "rawvideo", "av1", "cfhd", "hap"},
		Audio:       []string{"aac", "alac", "pcm_*", "mp3", "ac3", "eac3", "flac", "opus"},
		Subtitle:    []string{"mov_text"},
		Chapters:    true,
		PreferVideo: []string{"prores", "h264", "hevc"}, PreferAudio: []string{"pcm_s16le", "aac", "alac"}, PreferSubtitle: []string{"mov_text"},
		Args: []string{"-movflags", "+faststart"},
	},
	{
		ID: "webm", Label: "WebM", Ext: "webm", Muxer: "webm",
		Description: "The open web format: VP8, VP9 or AV1 video with Opus or Vorbis audio, and WebVTT subtitles.",
		Video:       []string{"vp9", "vp8", "av1"}, Audio: []string{"opus", "vorbis"}, Subtitle: []string{"webvtt"},
		Chapters:    true,
		PreferVideo: []string{"vp9", "av1"}, PreferAudio: []string{"opus"}, PreferSubtitle: []string{"webvtt"},
	},
	{
		ID: "ts", Label: "MPEG-TS", Ext: "ts", Muxer: "mpegts",
		Description: "Transport stream for broadcast and streaming: survives cut transfers, but no chapters or text subtitles.",
		Video:       []string{"h264", "hevc", "mpeg2video", "mpeg1video", "av1", "vvc"},
		Audio:       []string{"aac", "mp3", "mp2", "ac3", "eac3", "opus", "dts", "truehd"},
		Subtitle:    []string{"dvb_subtitle"},
		PreferVideo: []string{"h264", "hevc"}, PreferAudio: []string{"aac", "ac3"},
	},
	{
		ID: "avi", Label: "AVI", Ext: "avi", Muxer: "avi",
		Description: "The old Windows format, for legacy players and hardware. No subtitles or chapters.",
		Video:       []string{"mpeg4", "h264", "mjpeg", "msmpeg4v2", "msmpeg4v3", "mpeg2video", "huffyuv", "ffv1", "utvideo", "rawvideo", "ffvhuff"},
		Audio:       []string{"mp3", "ac3", "pcm_s16le", "pcm_u8", "mp2", "aac"},
		PreferVideo: []string{"mpeg4", "h264"}, PreferAudio: []string{"mp3", "ac3"},
	},
	{
		ID: "flv", Label: "FLV", Ext: "flv", Muxer: "flv",
		Description: "Flash Video, still what many RTMP streaming servers expect.",
		Video:       []string{"h264", "flv1", "hevc", "av1"}, Audio: []string{"aac", "mp3", "speex", "nellymoser"},
		PreferVideo: []string{"h264"}, PreferAudio: []string{"aac"},
	},
	{
		ID: "ogv", Label: "OGV", Ext: "ogv", Muxer: "ogg",
		Description: "Ogg video with Theora or VP8, fully patent-free.",
		Video:       []string{"theora", "vp8"}, Audio: []string{"vorbis", "opus", "flac", "speex"},
		Chapters:    true,
		PreferVideo: []string{"theora", "vp8"}, PreferAudio: []string{"vorbis", "opus"},
	},
	{
		ID: "gif", Label: "GIF", Ext: "gif", Muxer: "gif",
		Description: "Animated GIF: silent, 256 colors, large files. Lower the resolution and frame rate.",
		Video:       []string{"gif"},
		PreferVideo: []string{"gif"},
	},
	// Audio only.
	{
		ID: "mp3", Label: "MP3", Ext: "mp3", Muxer: "mp3", AudioOnly: true,
		Description: "MP3 audio, playable everywhere.",
		Audio:       []string{"mp3"}, PreferAudio: []string{"mp3"},
	},
	{
		ID: "m4a", Label: "M4A", Ext: "m4a", Muxer: "ipod", AudioOnly: true,
		Description: "AAC or Apple Lossless in an MPEG-4 file, the iTunes and Apple Music format.",
		Audio:       []string{"aac", "alac"}, Chapters: true, PreferAudio: []string{"aac", "alac"},
		Args: []string{"-movflags", "+faststart"},
	},
	{
		ID: "opus", Label: "Opus", Ext: "opus", Muxer: "opus", AudioOnly: true,
		Description: "Opus in Ogg: the best quality per bit for speech and music.",
		Audio:       []string{"opus"}, PreferAudio: []string{"opus"},
	},
	{
		ID: "ogg", Label: "OGG", Ext: "ogg", Muxer: "ogg", AudioOnly: true,
		Description: "Ogg audio with Vorbis, Opus or FLAC.",
		Audio:       []string{"vorbis", "opus", "flac", "speex"}, Chapters: true, PreferAudio: []string{"vorbis", "opus"},
	},
	{
		ID: "flac", Label: "FLAC", Ext: "flac", Muxer: "flac", AudioOnly: true,
		Description: "Lossless compressed audio: bit-perfect, about half the size of WAV.",
		Audio:       []string{"flac"}, PreferAudio: []string{"flac"},
	},
	{
		ID: "wav", Label: "WAV", Ext: "wav", Muxer: "wav", AudioOnly: true,
		Description: "Uncompressed PCM audio for editing and archiving.",
		Audio:       []string{"pcm_*", "adpcm_*"}, PreferAudio: []string{"pcm_s16le", "pcm_s24le"},
	},
	{
		ID: "aiff", Label: "AIFF", Ext: "aiff", Muxer: "aiff", AudioOnly: true,
		Description: "Apple's uncompressed audio format.",
		Audio:       []string{"pcm_s16be", "pcm_s24be", "pcm_s32be", "pcm_f32be"}, PreferAudio: []string{"pcm_s16be", "pcm_s24be"},
	},
	{
		ID: "aac", Label: "AAC", Ext: "aac", Muxer: "adts", AudioOnly: true,
		Description: "Raw AAC stream (ADTS), for players and devices that want plain AAC.",
		Audio:       []string{"aac"}, PreferAudio: []string{"aac"},
	},
	{
		ID: "ac3", Label: "AC-3", Ext: "ac3", Muxer: "ac3", AudioOnly: true,
		Description: "Dolby Digital surround audio.",
		Audio:       []string{"ac3"}, PreferAudio: []string{"ac3"},
	},
	{
		ID: "mka", Label: "MKA", Ext: "mka", Muxer: "matroska", AudioOnly: true,
		Description: "Matroska audio: any audio codec, several tracks and chapters.",
		Audio:       []string{"*"}, Chapters: true, PreferAudio: []string{"opus", "flac", "aac"},
	},
}
