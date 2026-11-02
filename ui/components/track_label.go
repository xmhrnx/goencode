package components

import (
	"strings"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// languages names the language codes discs and files tag tracks with.
var languages = map[string]string{
	"eng": "English", "fre": "French", "fra": "French", "ger": "German", "deu": "German",
	"spa": "Spanish", "ita": "Italian", "por": "Portuguese", "jpn": "Japanese", "chi": "Chinese",
	"zho": "Chinese", "kor": "Korean", "rus": "Russian", "dut": "Dutch", "nld": "Dutch",
	"swe": "Swedish", "dan": "Danish", "nor": "Norwegian", "fin": "Finnish", "pol": "Polish",
	"cze": "Czech", "ces": "Czech", "hun": "Hungarian", "tur": "Turkish", "gre": "Greek",
	"ell": "Greek", "heb": "Hebrew", "ara": "Arabic", "hin": "Hindi", "tha": "Thai",
	"ind": "Indonesian", "may": "Malay", "msa": "Malay", "vie": "Vietnamese", "ukr": "Ukrainian",
	"rum": "Romanian", "ron": "Romanian", "bul": "Bulgarian", "hrv": "Croatian", "slv": "Slovenian",
	"slo": "Slovak", "slk": "Slovak", "ice": "Icelandic", "isl": "Icelandic", "est": "Estonian",
	"lav": "Latvian", "lit": "Lithuanian", "per": "Persian", "fas": "Persian", "cat": "Catalan",
	"srp": "Serbian", "tam": "Tamil", "tel": "Telugu",
}

// Language names a track's language code: "eng" is "English".
func Language(code string) string {
	code = strings.ToLower(code)
	if code == "" || code == "und" {
		return "Unknown language"
	}
	if n, ok := languages[code]; ok {
		return n
	}
	return strings.ToUpper(code)
}

// TrackLabel describes a track, apart from its language: "TrueHD 7.1 ·
// Commentary".
func TrackLabel(st ffmpeg.Stream) string {
	codec := shortCodec(st.Codec)
	if st.Channels > 0 {
		codec += " " + channels(st.Channels)
	}
	parts := []string{codec}
	if st.Forced {
		parts = append(parts, "forced")
	}
	if st.Title != "" {
		parts = append(parts, st.Title)
	}
	return strings.Join(parts, " · ")
}
