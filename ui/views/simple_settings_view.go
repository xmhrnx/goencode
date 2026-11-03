package views

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// SimpleSettings is the quick mode: pick a preset, then nudge quality and
// size. Everything else follows the preset.
func SimpleSettings(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	presets := settings.AvailablePresets(s.Caps)

	var grid []giode.Element
	for i := 0; i < len(presets); i += 2 {
		pair := []giode.Element{}
		for _, p := range presets[i:min(i+2, len(presets))] {
			id := p.ID
			pair = append(pair, components.PresetCard(p, s.Prefs.Preset == id, func() { s.ApplyPreset(id) }))
		}
		if len(pair) == 1 {
			pair = append(pair, giode.Box(giode.Styles{}))
		}
		grid = append(grid, components.EqualRow(theme.S2, pair...))
	}

	tune := []giode.Element{}
	if f, ok := settings.QualityFraction(st, s.Caps); ok {
		tune = append(tune, giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
			components.InfoLabel("simple-quality", "Quality",
				"Moves "+s.VideoProfile().Label+"'s quality setting. Left makes smaller files that look softer; right keeps more detail and makes bigger files. The middle suits most videos."),
			components.Slider("simple-quality/"+st.Video.Encoder+"/"+st.Video.RateControl, f,
				components.SliderSpec{Min: 0, Max: 1, Step: 0.01, LeftLabel: "Smaller file", RightLabel: "Higher quality", Format: qualityWord},
				func(v float64) { s.Update(func(x *settings.Settings) { settings.SetQualityFraction(x, s.Caps, v) }) })))
	}
	if st.Video.Mode == settings.ModeEncode {
		resLabels, resValues := []string{}, []string{}
		for _, r := range resolutions {
			resLabels, resValues = append(resLabels, r.label), append(resValues, strconv.Itoa(r.height))
		}
		tune = append(tune, components.OptionRow(components.InfoLabel("simple-res", "Resolution", catalog.HelpResolution(s.VideoProfile().Label)),
			components.Select("simple-res", resLabels, resValues, strconv.Itoa(st.Video.Height),
				func(v string) { s.Update(func(x *settings.Settings) { x.Video.Height, _ = strconv.Atoi(v) }) })))
	}

	cards := []giode.Element{
		giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
			giode.Text("Presets", theme.Label),
			giode.Stack(giode.Styles{Gap: theme.S2, Width: -1}, grid...)),
		recipeSummary(env),
	}
	if len(tune) > 0 {
		cards = append(cards, components.SectionCard("tune", "Fine-tune", "Applies to every file", nil, tune...))
	}
	cards = append(cards, OutputSettings(env))
	return giode.Stack(giode.Styles{Gap: theme.S4, Width: -1}, cards...)
}

// recipeSummary shows what the current recipe does in a row of pills.
func recipeSummary(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	c := catalog.ContainerByID(st.Container)
	pills := []giode.Element{components.Pill(c.Label, "primary", "box")}
	switch st.Video.Mode {
	case settings.ModeCopy:
		pills = append(pills, components.Pill("Video copied", "success", "zap"))
	case settings.ModeEncode:
		label := catalog.Codec(st.Video.Codec, "").Label + " · " + s.VideoProfile().Label
		tone := "muted"
		if s.VideoProfile().Hardware != "" {
			tone = "success"
		}
		pills = append(pills, components.Pill(label, tone, "film"))
	}
	switch st.Audio.Mode {
	case settings.ModeCopy:
		pills = append(pills, components.Pill("Audio copied", "success", "zap"))
	case settings.ModeEncode:
		label := catalog.Codec(st.Audio.Codec, "").Label
		if rc := s.AudioProfile().RateControl(st.Audio.RateControl); rc != nil && rc.Bitrate {
			label += " · " + kbps(float64(st.Audio.Bitrate))
		}
		pills = append(pills, components.Pill(label, "muted", "music"))
	}
	return giode.Stack(theme.Card().With(giode.Styles{Gap: theme.S2, Width: -1}),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Recipe", theme.Label)),
			components.Button("customize", components.ButtonSpec{Label: "Customize", Icon: "tune", Kind: components.Ghost, Small: true,
				Tooltip: "Open the in-depth settings", OnClick: func() { s.SetAdvanced(true) }})),
		giode.Row(giode.Styles{Gap: 6, Width: -1}, pills...),
	)
}

func qualityWord(f float64) string {
	switch {
	case f < 0.2:
		return "Smallest"
	case f < 0.4:
		return "Compact"
	case f < 0.62:
		return "Balanced"
	case f < 0.82:
		return "High"
	default:
		return "Near lossless"
	}
}
