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

// VideoSettings is the advanced video tab: copy or encode, codec, encoder,
// quality, picture size and rate, and the encoder's own options.
func VideoSettings(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	c := catalog.ContainerByID(st.Container)
	if !c.Holds(ffmpeg.KindVideo) {
		return components.SectionCard("video", "Video", c.Label+" is an audio format", nil,
			components.Notice("info", "info", c.Label+" can't hold video, so the video is left out. Pick a video format to keep it."))
	}

	mode := components.Segmented("video-mode",
		[]string{"Encode", "Copy", "Off"}, []string{"encode", "copy", "drop"}, nil, string(st.Video.Mode),
		func(v string) { s.Update(func(x *settings.Settings) { x.Video.Mode = settings.StreamMode(v) }) })

	rows := []giode.Element{mode}
	switch st.Video.Mode {
	case settings.ModeCopy:
		rows = append(rows,
			components.Notice("success", "zap", "The video is copied bit for bit: no quality loss, and remuxing takes seconds."),
			components.OptionRow(components.InfoLabel("v-fallback", "Convert if it can't be copied", catalog.HelpCopyFallback("video")),
				components.Switch("v-fallback", st.Video.CopyFallback, func(on bool) { s.Update(func(x *settings.Settings) { x.Video.CopyFallback = on }) })))
		if st.Video.CopyFallback {
			rows = append(rows, codecRows(env)...)
		}
	case settings.ModeEncode:
		rows = append(rows, codecRows(env)...)
		rows = append(rows, qualityRows(env)...)
	default:
		rows = append(rows, components.Notice("info", "info", "No video in the output: only audio and subtitles are kept."))
	}

	cards := []giode.Element{components.SectionCard("video", "Video", "How the picture is stored", nil, rows...)}
	if st.Video.Mode == settings.ModeEncode {
		cards = append(cards, pictureCard(env), EncoderOptions(env, ffmpeg.KindVideo))
	}
	return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1}, cards...)
}

// codecRows are the codec and encoder dropdowns.
func codecRows(env *Env) []giode.Element {
	s := env.Store
	st := s.Settings
	labels, values := codecChoices(s, ffmpeg.KindVideo)
	encLabels, encValues := encoderChoices(s, st.Video.Codec, ffmpeg.KindVideo)
	codecName := catalog.Codec(st.Video.Codec, "").Label
	p := s.VideoProfile()
	rows := []giode.Element{
		components.OptionRow(components.InfoLabel("v-codec", "Codec", codecHelp(s, st.Video.Codec)),
			components.Select("v-codec", labels, values, st.Video.Codec, s.SetVideoCodec)),
		components.OptionRow(components.InfoLabel("v-encoder", "Encoder", p.Summary+" "+catalog.HelpEncoder(codecName)),
			components.Select("v-encoder", encLabels, encValues, st.Video.Encoder, s.SetVideoEncoder)),
	}
	if !s.HWReady {
		rows = append(rows, giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 12, Color: "theme.muted"}),
			giode.Text("Testing GPU encoders…", theme.Caption)))
	}
	return rows
}

// qualityRows are the rate control and its value.
func qualityRows(env *Env) []giode.Element {
	s := env.Store
	st := s.Settings
	p := s.VideoProfile()
	rc := p.RateControl(st.Video.RateControl)
	if rc == nil {
		return nil
	}
	labels, values := rateChoices(p)
	rows := []giode.Element{}
	if len(values) > 1 {
		rows = append(rows, components.OptionRow(components.InfoLabel("v-rc", "Rate control", catalog.HelpRateControl(p.Label)),
			components.Select("v-rc/"+st.Video.Encoder, labels, values, rc.ID, s.SetVideoRateControl)))
	}
	switch {
	case rc.None:
		rows = append(rows, components.Notice("info", "info", rc.Help))
	case rc.Bitrate:
		rows = append(rows,
			giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
				components.InfoLabel("v-bitrate", "Bitrate", rc.Help),
				components.Slider("v-bitrate/"+st.Video.Encoder, float64(st.Video.Bitrate),
					components.SliderSpec{Min: rc.Min, Max: min(rc.Max, 40000), Step: rc.Step, Format: kbps, LeftLabel: "Smaller file", RightLabel: "Higher quality"},
					func(v float64) { s.Update(func(x *settings.Settings) { x.Video.Bitrate = int(v) }) })),
			components.OptionRow(components.InfoLabel("v-maxrate", "Maximum bitrate",
				"Caps short peaks of the bitrate for streaming or devices with limited bandwidth. Empty: no cap."),
				components.Number("v-maxrate", intText(st.Video.MaxBitrate), "No limit", "kb/s", 0, 0,
					func(t string) { s.Update(func(x *settings.Settings) { x.Video.MaxBitrate, _ = strconv.Atoi(t) }) })))
	default:
		spec := rateSlider(rc)
		spec.Hint = markFor(rc, st.Video.Quality)
		rows = append(rows, giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
			components.InfoLabel("v-quality/"+st.Video.Encoder, "Quality", rc.Help),
			components.Slider("v-quality/"+st.Video.Encoder+"/"+rc.ID, st.Video.Quality, spec,
				func(v float64) { s.Update(func(x *settings.Settings) { x.Video.Quality = v }) })))
	}
	return rows
}

// pictureCard holds size, frame rate, pixel format and hardware decoding.
func pictureCard(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	enc := s.VideoProfile().Label

	resLabels, resValues := []string{}, []string{}
	for _, r := range resolutions {
		resLabels, resValues = append(resLabels, r.label), append(resValues, strconv.Itoa(r.height))
	}
	fpsLabels, fpsValues := []string{}, []string{}
	for _, f := range frameRates {
		fpsLabels, fpsValues = append(fpsLabels, f.label), append(fpsValues, strconv.FormatFloat(f.rate, 'f', -1, 64))
	}
	var formats []string
	if h := s.EncoderHelp(st.Video.Encoder); h != nil {
		formats = h.PixelFormats
	}
	pixLabels, pixValues := []string{"Automatic"}, []string{""}
	for _, f := range formats {
		pixLabels, pixValues = append(pixLabels, f), append(pixValues, f)
	}

	rows := []giode.Element{
		components.OptionRow(components.InfoLabel("v-res", "Resolution", catalog.HelpResolution(enc)),
			components.Select("v-res", resLabels, resValues, strconv.Itoa(st.Video.Height),
				func(v string) { s.Update(func(x *settings.Settings) { x.Video.Height, _ = strconv.Atoi(v) }) })),
		components.OptionRow(components.InfoLabel("v-fps", "Frame rate", catalog.HelpFrameRate(enc)),
			components.Select("v-fps", fpsLabels, fpsValues, strconv.FormatFloat(st.Video.FrameRate, 'f', -1, 64),
				func(v string) {
					s.Update(func(x *settings.Settings) { x.Video.FrameRate, _ = strconv.ParseFloat(v, 64) })
				})),
	}
	if len(formats) > 1 {
		rows = append(rows, components.OptionRow(components.InfoLabel("v-pix/"+st.Video.Encoder, "Pixel format", catalog.HelpPixelFormat(enc, formats)),
			components.Select("v-pix/"+st.Video.Encoder, pixLabels, pixValues, st.Video.PixelFormat,
				func(v string) { s.Update(func(x *settings.Settings) { x.Video.PixelFormat = v }) })))
	}
	rows = append(rows, components.OptionRow(components.InfoLabel("v-hwdec", "Hardware decoding", catalog.HelpHWDecode()),
		components.Switch("v-hwdec", st.Video.HWDecode, func(on bool) { s.Update(func(x *settings.Settings) { x.Video.HWDecode = on }) })))
	return components.SectionCard("image", "Picture", "Size, motion and color", nil, rows...)
}

func intText(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}
