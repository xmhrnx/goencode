package settings

import (
	"strconv"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// Normalize returns s adjusted to what the container and this ffmpeg
// support: an impossible codec is replaced with the container's preferred
// one, an unavailable encoder with the codec's default, an unknown rate
// control with the encoder's first, out-of-range values are clamped, and
// option values the encoder doesn't accept are cleared.
func Normalize(s Settings, caps *ffmpeg.Capabilities) Settings {
	s = s.Clone()
	c := catalog.ContainerByID(s.Container)
	s.Container = c.ID
	if caps == nil {
		return s
	}

	// Video.
	if !c.Holds(ffmpeg.KindVideo) {
		s.Video.Mode = ModeDrop
	} else if s.Video.Mode == "" {
		s.Video.Mode = ModeEncode
	}
	if s.Video.Mode == ModeEncode || s.Video.CopyFallback {
		s.Video.Codec, s.Video.Encoder = fixCodec(caps, c, ffmpeg.KindVideo, s.Video.Codec, s.Video.Encoder)
		if s.Video.Codec == "" && s.Video.Mode == ModeEncode {
			s.Video.Mode = ModeCopy
		}
	}
	if s.Video.Encoder != "" {
		p := ProfileOf(caps, s.Video.Encoder, ffmpeg.KindVideo)
		s.Video.RateControl, s.Video.Quality, s.Video.Bitrate = fixRate(p, s.Video.RateControl, s.Video.Quality, s.Video.Bitrate)
		fixOptions(p, s.Video.Options[s.Video.Encoder])
		if h := caps.CachedEncoderHelp(s.Video.Encoder); h != nil && s.Video.PixelFormat != "" && len(h.PixelFormats) > 0 && !contains(h.PixelFormats, s.Video.PixelFormat) {
			s.Video.PixelFormat = ""
		}
	}

	// Audio.
	if !c.Holds(ffmpeg.KindAudio) {
		s.Audio.Mode = ModeDrop
	} else if s.Audio.Mode == "" {
		s.Audio.Mode = ModeEncode
	}
	if s.Audio.Tracks == "" {
		s.Audio.Tracks = TracksAll
	}
	if s.Audio.Mode == ModeEncode || s.Audio.CopyFallback {
		s.Audio.Codec, s.Audio.Encoder = fixCodec(caps, c, ffmpeg.KindAudio, s.Audio.Codec, s.Audio.Encoder)
		if s.Audio.Codec == "" && s.Audio.Mode == ModeEncode {
			s.Audio.Mode = ModeCopy
		}
	}
	if s.Audio.Encoder != "" {
		p := ProfileOf(caps, s.Audio.Encoder, ffmpeg.KindAudio)
		s.Audio.RateControl, s.Audio.Quality, s.Audio.Bitrate = fixRate(p, s.Audio.RateControl, s.Audio.Quality, s.Audio.Bitrate)
		fixOptions(p, s.Audio.Options[s.Audio.Encoder])
		if h := caps.CachedEncoderHelp(s.Audio.Encoder); h != nil && s.Audio.SampleRate != 0 && len(h.SampleRates) > 0 &&
			!contains(h.SampleRates, strconv.Itoa(s.Audio.SampleRate)) {
			s.Audio.SampleRate = 0
		}
	}

	// Subtitles.
	if s.Subtitles.Mode == "" {
		s.Subtitles.Mode = SubsKeep
	}
	if s.Subtitles.Tracks == "" {
		s.Subtitles.Tracks = TracksAll
	}
	if s.Subtitles.Mode == SubsBurn && s.Video.Mode != ModeEncode {
		s.Subtitles.Mode = SubsKeep
	}
	if s.Subtitles.Codec != "" && !contains(TextSubtitleCodecs(caps, c), s.Subtitles.Codec) {
		s.Subtitles.Codec = ""
	}
	return s
}

func fixCodec(caps *ffmpeg.Capabilities, c *catalog.Container, kind ffmpeg.MediaKind, codec, encoder string) (string, string) {
	if !contains(Codecs(caps, c, kind), codec) {
		codec = PreferredCodec(caps, c, kind)
		encoder = ""
	}
	if codec == "" {
		return "", ""
	}
	if !contains(Encoders(caps, codec), encoder) {
		encoder = DefaultEncoder(caps, codec)
	}
	return codec, encoder
}

func fixRate(p *catalog.Profile, id string, q float64, br int) (string, float64, int) {
	rc := p.RateControl(id)
	if rc == nil {
		return "", q, br
	}
	if rc.ID != id {
		return rc.ID, rc.Default, int(rc.Default)
	}
	if !rc.Bitrate && !rc.None && (q < rc.Min || q > rc.Max) {
		q = rc.Default
	}
	if rc.Bitrate && (br <= 0 || float64(br) < rc.Min || float64(br) > rc.Max) {
		br = int(rc.Default)
	}
	return rc.ID, q, br
}

func fixOptions(p *catalog.Profile, opts map[string]string) {
	for k, v := range opts {
		o := p.Option(k)
		if o == nil || o.Kind != catalog.OptionChoice {
			continue
		}
		valid := false
		for _, c := range o.Choices {
			if c.Value == v {
				valid = true
			}
		}
		if !valid {
			delete(opts, k)
		}
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
