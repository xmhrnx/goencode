package store

import (
	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/settings"
)

// Update changes the recipe and keeps it valid for the container and
// this ffmpeg. Every settings control in the UI goes through it.
func (s *Store) Update(fn func(*settings.Settings)) {
	next := s.Settings.Clone()
	fn(&next)
	s.Settings = settings.Normalize(next, s.Caps)
	s.Prefs.Preset = ""
	s.loadHelp()
	s.SavePrefs()
}

// SetContainer switches the output format.
func (s *Store) SetContainer(id string) {
	s.Update(func(st *settings.Settings) { st.Container = id })
}

// SetVideoCodec switches the video codec, starting from its default
// encoder and that encoder's default quality.
func (s *Store) SetVideoCodec(codec string) {
	s.Update(func(st *settings.Settings) {
		st.Video.Codec = codec
		s.resetVideoEncoder(st, settings.DefaultEncoder(s.Caps, codec))
	})
}

// SetVideoEncoder switches encoder within the codec.
func (s *Store) SetVideoEncoder(enc string) {
	s.Update(func(st *settings.Settings) { s.resetVideoEncoder(st, enc) })
}

func (s *Store) resetVideoEncoder(st *settings.Settings, enc string) {
	st.Video.Encoder = enc
	p := settings.ProfileOf(s.Caps, enc, ffmpeg.KindVideo)
	if rc := p.RateControl(""); rc != nil {
		st.Video.RateControl, st.Video.Quality, st.Video.Bitrate = rc.ID, rc.Default, int(rc.Default)
	}
	st.Video.PixelFormat = ""
}

// SetVideoRateControl switches the rate control, starting at its default.
func (s *Store) SetVideoRateControl(id string) {
	s.Update(func(st *settings.Settings) {
		rc := settings.ProfileOf(s.Caps, st.Video.Encoder, ffmpeg.KindVideo).RateControl(id)
		st.Video.RateControl, st.Video.Quality, st.Video.Bitrate = rc.ID, rc.Default, int(rc.Default)
	})
}

// SetAudioCodec switches the audio codec.
func (s *Store) SetAudioCodec(codec string) {
	s.Update(func(st *settings.Settings) {
		st.Audio.Codec = codec
		s.resetAudioEncoder(st, settings.DefaultEncoder(s.Caps, codec))
	})
}

// SetAudioEncoder switches encoder within the codec.
func (s *Store) SetAudioEncoder(enc string) {
	s.Update(func(st *settings.Settings) { s.resetAudioEncoder(st, enc) })
}

func (s *Store) resetAudioEncoder(st *settings.Settings, enc string) {
	st.Audio.Encoder = enc
	p := settings.ProfileOf(s.Caps, enc, ffmpeg.KindAudio)
	if rc := p.RateControl(""); rc != nil {
		st.Audio.RateControl, st.Audio.Quality, st.Audio.Bitrate = rc.ID, rc.Default, int(rc.Default)
	}
	st.Audio.SampleRate = 0
}

// SetAudioRateControl switches the audio rate control.
func (s *Store) SetAudioRateControl(id string) {
	s.Update(func(st *settings.Settings) {
		rc := settings.ProfileOf(s.Caps, st.Audio.Encoder, ffmpeg.KindAudio).RateControl(id)
		st.Audio.RateControl, st.Audio.Quality, st.Audio.Bitrate = rc.ID, rc.Default, int(rc.Default)
	})
}

// ApplyPreset replaces the recipe with a preset, keeping output settings.
func (s *Store) ApplyPreset(id string) {
	for _, p := range settings.AvailablePresets(s.Caps) {
		if p.ID != id {
			continue
		}
		next := p.Build(s.Caps)
		next.Output = s.Settings.Output
		s.Settings = settings.Normalize(next, s.Caps)
		s.Prefs.Preset = id
		s.loadHelp()
		s.SavePrefs()
		return
	}
}

// SetAdvanced switches between the simple and advanced settings.
func (s *Store) SetAdvanced(on bool) {
	s.Prefs.Advanced = on
	s.SavePrefs()
}

// VideoProfile and AudioProfile describe the current encoders.
func (s *Store) VideoProfile() *catalog.Profile {
	return settings.ProfileOf(s.Caps, s.Settings.Video.Encoder, ffmpeg.KindVideo)
}

func (s *Store) AudioProfile() *catalog.Profile {
	return settings.ProfileOf(s.Caps, s.Settings.Audio.Encoder, ffmpeg.KindAudio)
}

// EncoderHelp returns ffmpeg's help for an encoder if it has been loaded;
// it's loaded in the background whenever the encoder is chosen.
func (s *Store) EncoderHelp(enc string) *ffmpeg.EncoderHelp {
	if s.Caps == nil || enc == "" {
		return nil
	}
	return s.Caps.CachedEncoderHelp(enc)
}

// loadHelp fetches the help of the chosen encoders, for their full option
// lists, pixel formats and sample rates.
func (s *Store) loadHelp() {
	if s.Caps == nil {
		return
	}
	for _, enc := range []string{s.Settings.Video.Encoder, s.Settings.Audio.Encoder} {
		if enc == "" || s.Caps.CachedEncoderHelp(enc) != nil {
			continue
		}
		caps := s.Caps
		go func(enc string) {
			caps.EncoderHelp(s.ctx, enc)
			s.host.Post(func() { s.Settings = settings.Normalize(s.Settings, caps) })
		}(enc)
	}
}
