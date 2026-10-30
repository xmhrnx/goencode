package catalog

import "strings"

// CodecInfo is a codec's display name and the explanation shown on hover.
type CodecInfo struct {
	Label string
	Help  string
	// Rank orders popular codecs first in lists; 0 is unranked.
	Rank int
}

var codecInfo = map[string]CodecInfo{
	"h264":       {"H.264 / AVC", "The most compatible video codec: plays on every phone, browser, TV and editor.", 1},
	"hevc":       {"H.265 / HEVC", "About half the size of H.264 at the same quality. Supports 10-bit and HDR. Plays on modern devices; not in Firefox.", 2},
	"av1":        {"AV1", "The most efficient royalty-free codec: around 30% smaller than HEVC. Plays in browsers and recent devices; slower to encode.", 3},
	"vp9":        {"VP9", "Google's royalty-free codec used by YouTube. Similar efficiency to HEVC and plays in every browser.", 4},
	"vvc":        {"H.266 / VVC", "The successor to HEVC. Very efficient, but almost nothing plays it yet.", 9},
	"prores":     {"Apple ProRes", "An editing codec: very high quality and fast to scrub, with large files. Native in Final Cut Pro.", 5},
	"dnxhd":      {"Avid DNxHD / DNxHR", "Avid's editing codec, an alternative to ProRes for Media Composer and Resolve.", 6},
	"ffv1":       {"FFV1 (lossless)", "Mathematically lossless video for archiving.", 7},
	"vp8":        {"VP8", "The first WebM codec. Superseded by VP9.", 10},
	"mpeg4":      {"MPEG-4 Part 2", "DivX/Xvid-era codec for old players.", 11},
	"mpeg2video": {"MPEG-2", "DVD and broadcast video.", 12},
	"mjpeg":      {"Motion JPEG", "Every frame a JPEG: easy to edit, large files.", 13},
	"theora":     {"Theora", "The original Ogg video codec.", 14},
	"gif":        {"GIF", "Animated GIF images.", 15},
	"utvideo":    {"Ut Video (lossless)", "Fast lossless intermediate codec.", 16},
	"huffyuv":    {"HuffYUV (lossless)", "Fast lossless codec for capture.", 17},
	"rawvideo":   {"Uncompressed", "Raw video frames. Enormous files.", 30},

	"aac":       {"AAC", "The audio of MP4 and streaming: good quality from 128 kb/s, plays everywhere.", 1},
	"opus":      {"Opus", "The best lossy audio codec at every bitrate, for music and speech. Native to WebM, MKV and Ogg.", 2},
	"mp3":       {"MP3", "Universally compatible lossy audio.", 3},
	"flac":      {"FLAC (lossless)", "Lossless compression at about half the size of WAV.", 4},
	"ac3":       {"Dolby Digital (AC-3)", "Surround sound for DVDs, Blu-rays and AV receivers.", 5},
	"eac3":      {"Dolby Digital Plus", "More efficient surround sound used by streaming services.", 6},
	"alac":      {"Apple Lossless", "Lossless audio for the Apple ecosystem.", 7},
	"vorbis":    {"Vorbis", "Open lossy audio, the predecessor of Opus.", 8},
	"pcm_s16le": {"PCM 16-bit", "Uncompressed CD-quality audio.", 9},
	"pcm_s24le": {"PCM 24-bit", "Uncompressed studio-resolution audio.", 10},
	"pcm_s16be": {"PCM 16-bit (BE)", "Uncompressed audio for AIFF.", 11},
	"pcm_s24be": {"PCM 24-bit (BE)", "Uncompressed 24-bit audio for AIFF.", 12},
	"mp2":       {"MPEG Layer II", "Broadcast audio.", 13},
	"dts":       {"DTS", "Cinema surround format (ffmpeg's encoder is experimental).", 20},
	"truehd":    {"Dolby TrueHD", "Lossless surround of Blu-ray (experimental encoder).", 21},

	"subrip":            {"SubRip (SRT)", "Plain text subtitles with timing, the most compatible text format.", 1},
	"ass":               {"Advanced SubStation (ASS)", "Text subtitles with fonts, colors and positioning, popular for anime.", 2},
	"ssa":               {"SubStation Alpha", "The older version of ASS.", 3},
	"webvtt":            {"WebVTT", "The web's subtitle format, required by WebM.", 4},
	"mov_text":          {"MP4 Timed Text", "The text subtitle format of MP4 and MOV.", 5},
	"dvd_subtitle":      {"VobSub (DVD)", "Picture-based DVD subtitles.", 6},
	"hdmv_pgs_subtitle": {"PGS (Blu-ray)", "Picture-based Blu-ray subtitles.", 7},
	"dvb_subtitle":      {"DVB subtitles", "Picture-based broadcast subtitles.", 8},
	"text":              {"Raw text", "Untimed plain text.", 9},
}

// Codec returns a codec's display information, falling back to ffmpeg's
// description.
func Codec(id, ffmpegDesc string) CodecInfo {
	if ci, ok := codecInfo[id]; ok {
		return ci
	}
	label := ffmpegDesc
	if label == "" {
		label = strings.ToUpper(id)
	}
	return CodecInfo{Label: label, Help: ffmpegDesc + " (" + id + "). Encoded with ffmpeg's defaults unless you set options below.", Rank: 100}
}
