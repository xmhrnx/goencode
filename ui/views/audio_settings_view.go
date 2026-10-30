package views

import (
	"strconv"

	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// AudioSettings is the advanced audio tab: which tracks, copy or encode,
// codec, encoder, bitrate or quality, channels, sample rate and loudness.
func AudioSettings(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	c := catalog.ContainerByID(st.Container)
	if !c.Holds(ffmpeg.KindAudio) {
		return components.SectionCard("music", "Audio", c.Label+" has no sound", nil,
			components.Notice("info", "info", c.Label+" can't hold audio, so it's left out."))
	}

	mode := components.Segmented("audio-mode",
		[]string{"Encode", "Copy", "Off"}, []string{"encode", "copy", "drop"}, nil, string(st.Audio.Mode),
		func(v string) { s.Update(func(x *settings.Settings) { x.Audio.Mode = settings.StreamMode(v) }) })

	rows := []giode.Element{mode}
	if st.Audio.Mode != settings.ModeDrop {
		rows = append(rows, trackRows(env, "a", st.Audio.Tracks, st.Audio.Languages,
			func(r settings.TrackRule) { s.Update(func(x *settings.Settings) { x.Audio.Tracks = r }) },
			func(l string) { s.Update(func(x *settings.Settings) { x.Audio.Languages = l }) })...)
	}
	switch st.Audio.Mode {
	case settings.ModeCopy:
		rows = append(rows,
			components.Notice("success", "zap", "Audio is copied untouched."),
			components.OptionRow(components.InfoLabel("a-fallback", "Convert if it can't be copied", catalog.HelpCopyFallback("audio")),
				components.Switch("a-fallback", st.Audio.CopyFallback, func(on bool) { s.Update(func(x *settings.Settings) { x.Audio.CopyFallback = on }) })))
		if st.Audio.CopyFallback {
			rows = append(rows, audioCodecRows(env)...)
		}
	case settings.ModeEncode:
		rows = append(rows, audioCodecRows(env)...)
		rows = append(rows, audioQualityRows(env)...)
	default:
		rows = append(rows, components.Notice("info", "info", "The output will be silent."))
	}

	cards := []giode.Element{components.SectionCard("music", "Audio", "Tracks, codec and sound", nil, rows...)}
	if st.Audio.Mode == settings.ModeEncode {
		cards = append(cards, soundCard(env), EncoderOptions(env, ffmpeg.KindAudio))
	}
	return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1}, cards...)
}

// trackRows picks all tracks, the first one, or tracks by language.
func trackRows(env *Env, prefix string, rule settings.TrackRule, langs string, onRule func(settings.TrackRule), onLangs func(string)) []giode.Element {
	kind := "audio"
	if prefix == "s" {
		kind = "subtitle"
	}
	rows := []giode.Element{
		components.OptionRow(components.InfoLabel(prefix+"-tracks", "Tracks", catalog.HelpTrackSelection(kind)),
			giode.Box(giode.Styles{Width: 210}, components.Segmented(prefix+"-tracks", []string{"All", "First", "Language"},
				[]string{"all", "first", "languages"}, nil, string(rule), func(v string) { onRule(settings.TrackRule(v)) }))),
	}
	if rule == settings.TracksLanguages {
		rows = append(rows, components.OptionRow(
			components.InfoLabel(prefix+"-langs", "Languages", "Language codes to keep, separated by commas, as tagged in the file: eng, jpn, fre. Files without a matching track keep their first one."),
			components.TextInput(prefix+"-langs", langs, "eng, jpn", onLangs)))
	}
	return rows
}

func audioCodecRows(env *Env) []giode.Element {
	s := env.Store
	st := s.Settings
	labels, values := codecChoices(s, ffmpeg.KindAudio)
	encLabels, encValues := encoderChoices(s, st.Audio.Codec, ffmpeg.KindAudio)
	p := s.AudioProfile()
	return []giode.Element{
		components.OptionRow(components.InfoLabel("a-codec", "Codec", codecHelp(s, st.Audio.Codec)),
			components.Select("a-codec", labels, values, st.Audio.Codec, s.SetAudioCodec)),
		components.OptionRow(components.InfoLabel("a-encoder", "Encoder", p.Summary+" "+catalog.HelpEncoder(catalog.Codec(st.Audio.Codec, "").Label)),
			components.Select("a-encoder", encLabels, encValues, st.Audio.Encoder, s.SetAudioEncoder)),
	}
}

func audioQualityRows(env *Env) []giode.Element {
	s := env.Store
	st := s.Settings
	p := s.AudioProfile()
	rc := p.RateControl(st.Audio.RateControl)
	if rc == nil {
		return nil
	}
	var rows []giode.Element
	if labels, values := rateChoices(p); len(values) > 1 {
		rows = append(rows, components.OptionRow(components.InfoLabel("a-rc", "Rate control", catalog.HelpRateControl(p.Label)),
			components.Select("a-rc/"+st.Audio.Encoder, labels, values, rc.ID, s.SetAudioRateControl)))
	}
	switch {
	case rc.None:
		rows = append(rows, components.Notice("info", "info", rc.Help))
	case rc.Bitrate:
		rows = append(rows, giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
			components.InfoLabel("a-bitrate/"+st.Audio.Encoder, "Bitrate", rc.Help),
			components.Slider("a-bitrate/"+st.Audio.Encoder, float64(st.Audio.Bitrate),
				components.SliderSpec{Min: rc.Min, Max: min(rc.Max, 640), Step: rc.Step, Format: kbps, LeftLabel: "Smaller file", RightLabel: "Higher quality",
					Hint: markFor(rc, float64(st.Audio.Bitrate))},
				func(v float64) { s.Update(func(x *settings.Settings) { x.Audio.Bitrate = int(v) }) })))
	default:
		spec := rateSlider(rc)
		spec.Hint = markFor(rc, st.Audio.Quality)
		rows = append(rows, giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
			components.InfoLabel("a-quality/"+st.Audio.Encoder, "Quality", rc.Help),
			components.Slider("a-quality/"+st.Audio.Encoder+"/"+rc.ID, st.Audio.Quality, spec,
				func(v float64) { s.Update(func(x *settings.Settings) { x.Audio.Quality = v }) })))
	}
	return rows
}

// soundCard holds channels, sample rate and loudness normalization.
func soundCard(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	enc := s.AudioProfile().Label

	chLabels, chValues := []string{}, []string{}
	for _, c := range channelChoices {
		chLabels, chValues = append(chLabels, c.label), append(chValues, strconv.Itoa(c.n))
	}
	srLabels, srValues := []string{"Same as source"}, []string{"0"}
	rates := []string{"8000", "16000", "22050", "24000", "32000", "44100", "48000", "88200", "96000", "192000"}
	if h := s.EncoderHelp(st.Audio.Encoder); h != nil && len(h.SampleRates) > 0 {
		rates = h.SampleRates
	}
	for _, r := range rates {
		n, _ := strconv.Atoi(r)
		srLabels = append(srLabels, strconv.FormatFloat(float64(n)/1000, 'f', -1, 64)+" kHz")
		srValues = append(srValues, r)
	}

	return components.SectionCard("speaker", "Sound", "Channels, sample rate and loudness", nil,
		components.OptionRow(components.InfoLabel("a-ch", "Channels", catalog.HelpChannels(enc)),
			components.Select("a-ch", chLabels, chValues, strconv.Itoa(st.Audio.Channels),
				func(v string) { s.Update(func(x *settings.Settings) { x.Audio.Channels, _ = strconv.Atoi(v) }) })),
		components.OptionRow(components.InfoLabel("a-sr/"+st.Audio.Encoder, "Sample rate", catalog.HelpSampleRate(enc)),
			components.Select("a-sr/"+st.Audio.Encoder, srLabels, srValues, strconv.Itoa(st.Audio.SampleRate),
				func(v string) { s.Update(func(x *settings.Settings) { x.Audio.SampleRate, _ = strconv.Atoi(v) }) })),
		components.OptionRow(components.InfoLabel("a-loud", "Normalize loudness", catalog.HelpLoudness()),
			components.Switch("a-loud", st.Audio.Loudnorm, func(on bool) { s.Update(func(x *settings.Settings) { x.Audio.Loudnorm = on }) })),
	)
}
