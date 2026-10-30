package catalog

import (
	"fmt"
	"strings"
)

// Help texts for the settings shared by every encoder. They take the
// encoder's name so the explanation reads in the context of the chosen
// codec, as the per-encoder option texts do.

func HelpEncoder(codec string) string {
	return "The program that produces " + codec + ". Hardware encoders (marked with a chip) are much faster; software encoders compress better."
}

func HelpRateControl(enc string) string {
	return "How " + enc + " decides how many bits each part of the video gets: a constant quality target, or an average bitrate."
}

func HelpResolution(enc string) string {
	return "Scales the video before " + enc + " encodes it. Height is set and width follows the aspect ratio; smaller videos are never enlarged."
}

func HelpFrameRate(enc string) string {
	return "Output frames per second. Keep the source rate unless a device needs a specific one; halving it saves bits in " + enc + " but makes motion less smooth."
}

func HelpPixelFormat(enc string, formats []string) string {
	s := "Color sampling and bit depth " + enc + " encodes. yuv420p (8-bit) plays everywhere; 10-bit (…p10le) reduces banding and is needed for HDR."
	if len(formats) > 0 {
		shown := formats
		if len(shown) > 8 {
			shown = shown[:8]
		}
		s += " " + enc + " accepts: " + strings.Join(shown, ", ")
		if len(formats) > len(shown) {
			s += fmt.Sprintf(" and %d more", len(formats)-len(shown))
		}
		s += "."
	}
	return s
}

func HelpHWDecode() string {
	return "Decodes the source on the GPU when possible, freeing the CPU for encoding. Falls back to software automatically."
}

func HelpChannels(enc string) string {
	return "Number of audio channels " + enc + " writes. Downmixing 5.1 to stereo makes files smaller and suits headphones and phones."
}

func HelpSampleRate(enc string) string {
	return "Samples per second. 48 kHz is the video standard and 44.1 kHz the CD standard; keep the source unless " + enc + " or a device requires otherwise."
}

func HelpLoudness() string {
	return "Normalizes loudness to -16 LUFS with EBU R128 (loudnorm), so quiet and loud files play at a similar volume. Requires re-encoding."
}

func HelpCopy(kind string) string {
	return "Copies the " + kind + " without re-encoding: instant and lossless. If a track's codec can't be stored in the chosen container, it's converted instead (or dropped when that's off)."
}

func HelpBitrate(enc string) string {
	return "Average kilobits per second " + enc + " targets. Doubling it roughly doubles the file size."
}

func HelpTrackSelection(kind string) string {
	return "Which " + kind + " tracks of each file to keep. Language codes are ISO 639 (eng, fra, jpn); tracks without a language tag are kept when 'und' is listed."
}

func HelpCopyFallback(kind string) string {
	return "When copying, " + kind + " tracks whose codec the container can't hold are converted to its preferred codec. Off: they're dropped and reported."
}

func HelpSubtitleMode() string {
	return "Keep subtitles as separate tracks, convert them to the container's format, burn one into the picture, or remove them."
}

func HelpBurn() string {
	return "Draws the chosen subtitle track into the video itself so it shows on any player. Requires re-encoding the video and can't be turned off later."
}

func HelpChapters() string {
	return "Chapter markers let players jump between scenes. Not every container stores them."
}

func HelpMetadata() string {
	return "Title, artist, encoder tags and other metadata. Strip it to remove personal or leftover information."
}

func HelpOutputFolder() string {
	return "Where finished files go. 'Next to source' writes each file beside its original."
}

func HelpSuffix() string {
	return "Added to each file name so outputs never overwrite their source. Leave empty to keep the name (a different extension still avoids clashes)."
}

func HelpOverwrite() string {
	return "Replace files that already exist at the output path. Off: GoEncode numbers the new file instead."
}

func HelpContainer(c *Container) string {
	return c.Description
}
