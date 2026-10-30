package settings

import (
	"sort"
	"strings"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// Codecs lists the codecs of a kind that this ffmpeg can encode and the
// container can hold, popular ones first. Codecs the container rejects are
// left out, so the UI never offers an impossible combination.
func Codecs(caps *ffmpeg.Capabilities, c *catalog.Container, kind ffmpeg.MediaKind) []string {
	if caps == nil {
		return nil
	}
	var ids []string
	for _, cd := range caps.CodecsOfKind(kind) {
		if !c.Accepts(kind, cd.ID) || len(Encoders(caps, cd.ID)) == 0 {
			continue
		}
		ids = append(ids, cd.ID)
	}
	sort.Slice(ids, func(i, j int) bool {
		ri := catalog.Codec(ids[i], "").Rank
		rj := catalog.Codec(ids[j], "").Rank
		if ri != rj {
			return ri < rj
		}
		return ids[i] < ids[j]
	})
	return ids
}

// preferredEncoders orders software encoders for codecs with several.
var preferredEncoders = map[string][]string{
	"h264":   {"libx264"},
	"hevc":   {"libx265"},
	"av1":    {"libsvtav1", "libaom-av1", "librav1e"},
	"vp9":    {"libvpx-vp9"},
	"vp8":    {"libvpx"},
	"prores": {"prores_ks", "prores_aw", "prores"},
	"aac":    {"libfdk_aac", "aac"},
	"opus":   {"libopus", "opus"},
	"mp3":    {"libmp3lame", "libshine"},
	"vorbis": {"libvorbis", "vorbis"},
	"subrip": {"srt", "subrip"},
}

// skipEncoders are variants that only confuse a list of choices.
var skipEncoders = map[string]bool{"libx264rgb": true, "vorbis": true, "dvdsub": false}

// Encoders lists the usable encoders of a codec: software first (best
// first), then hardware encoders that passed their test.
func Encoders(caps *ffmpeg.Capabilities, codec string) []string {
	cd := caps.Codecs[codec]
	if cd == nil {
		return nil
	}
	var sw, hw []string
	for _, e := range cd.Encoders {
		if skipEncoders[e] || !caps.EncoderUsable(e) {
			continue
		}
		if _, isHW := ffmpeg.HardwareOf(e); isHW {
			hw = append(hw, e)
		} else {
			sw = append(sw, e)
		}
	}
	rank := func(e string) int {
		for i, p := range preferredEncoders[codec] {
			if p == e {
				return i
			}
		}
		return 99
	}
	sort.SliceStable(sw, func(i, j int) bool { return rank(sw[i]) < rank(sw[j]) })
	sort.Strings(hw)
	return append(sw, hw...)
}

// DefaultEncoder picks the encoder used when a codec is first chosen.
func DefaultEncoder(caps *ffmpeg.Capabilities, codec string) string {
	if list := Encoders(caps, codec); len(list) > 0 {
		return list[0]
	}
	return ""
}

// FastestEncoder prefers a working hardware encoder, for the "GPU" preset.
func FastestEncoder(caps *ffmpeg.Capabilities, codec string) string {
	list := Encoders(caps, codec)
	for _, fam := range []string{"_nvenc", "_qsv", "_amf", "_videotoolbox", "_vaapi", "_mediacodec"} {
		for _, e := range list {
			if strings.HasSuffix(e, fam) {
				return e
			}
		}
	}
	if len(list) > 0 {
		return list[0]
	}
	return ""
}

// PreferredCodec picks a codec for a container: its preferred list first,
// then whatever it accepts.
func PreferredCodec(caps *ffmpeg.Capabilities, c *catalog.Container, kind ffmpeg.MediaKind) string {
	available := Codecs(caps, c, kind)
	prefer := c.PreferVideo
	switch kind {
	case ffmpeg.KindAudio:
		prefer = c.PreferAudio
	case ffmpeg.KindSubtitle:
		prefer = c.PreferSubtitle
	}
	for _, p := range prefer {
		for _, a := range available {
			if a == p {
				return a
			}
		}
	}
	if len(available) > 0 {
		return available[0]
	}
	return ""
}

// ProfileOf returns the catalog profile of an encoder.
func ProfileOf(caps *ffmpeg.Capabilities, encoder string, kind ffmpeg.MediaKind) *catalog.Profile {
	desc := ""
	if caps != nil {
		if cd := caps.Codecs[caps.EncoderCodec[encoder]]; cd != nil {
			desc = cd.Description
		}
	}
	return catalog.ProfileFor(encoder, kind, desc)
}

// TextSubtitleCodecs lists the text subtitle codecs a container takes that
// this ffmpeg can write.
func TextSubtitleCodecs(caps *ffmpeg.Capabilities, c *catalog.Container) []string {
	var out []string
	for _, id := range Codecs(caps, c, ffmpeg.KindSubtitle) {
		switch id {
		case "subrip", "ass", "ssa", "webvtt", "mov_text", "text":
			out = append(out, id)
		}
	}
	return out
}
