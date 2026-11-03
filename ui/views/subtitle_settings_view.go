package views

import (
	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// SubtitleSettings keeps, burns in or drops subtitles, and picks the format
// text subtitles are stored in.
func SubtitleSettings(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	c := catalog.ContainerByID(st.Container)
	canKeep := c.Holds(ffmpeg.KindSubtitle)
	canBurn := c.Holds(ffmpeg.KindVideo) && st.Video.Mode == settings.ModeEncode

	labels, values := []string{}, []string{}
	if canKeep {
		labels, values = append(labels, "Keep"), append(values, string(settings.SubsKeep))
	}
	if canBurn {
		labels, values = append(labels, "Burn in"), append(values, string(settings.SubsBurn))
	}
	labels, values = append(labels, "Off"), append(values, string(settings.SubsDrop))

	var rows []giode.Element
	if len(values) > 1 {
		rows = append(rows, components.OptionRow(components.InfoLabel("s-mode", "Subtitles", catalog.HelpSubtitleMode()),
			giode.Box(giode.Styles{Width: -1}, components.Segmented("s-mode/"+st.Container+boolKey(canBurn), labels, values, nil, string(st.Subtitles.Mode),
				func(v string) { s.Update(func(x *settings.Settings) { x.Subtitles.Mode = settings.SubtitleMode(v) }) }))))
	}

	switch st.Subtitles.Mode {
	case settings.SubsKeep:
		if !canKeep {
			break
		}
		rows = append(rows, trackRows(env, "s", st.Subtitles.Tracks, st.Subtitles.Languages,
			func(r settings.TrackRule) { s.Update(func(x *settings.Settings) { x.Subtitles.Tracks = r }) },
			func(l string) { s.Update(func(x *settings.Settings) { x.Subtitles.Languages = l }) })...)
		if codecs := settings.TextSubtitleCodecs(s.Caps, c); len(codecs) > 0 {
			fl, fv := []string{"Keep when possible"}, []string{""}
			for _, id := range codecs {
				fl, fv = append(fl, catalog.Codec(id, "").Label), append(fv, id)
			}
			rows = append(rows, components.OptionRow(
				components.InfoLabel("s-codec", "Text format", "Converts text subtitles to this format. \"Keep when possible\" copies them and converts only those "+c.Label+" can't store. Picture subtitles (Blu-ray, DVD) can't become text and are dropped if the format can't hold them."),
				components.Select("s-codec/"+st.Container, fl, fv, st.Subtitles.Codec,
					func(v string) { s.Update(func(x *settings.Settings) { x.Subtitles.Codec = v }) })))
		}
	case settings.SubsBurn:
		rows = append(rows,
			components.Notice("info", "captions", "One track is drawn onto the picture, so it shows in every player and can't be turned off."),
			components.OptionRow(components.InfoLabel("s-burn", "Language to burn", catalog.HelpBurn()),
				components.TextInput("s-burn", st.Subtitles.BurnLanguage, "Forced or default track",
					func(v string) { s.Update(func(x *settings.Settings) { x.Subtitles.BurnLanguage = v }) })))
	default:
		if !canKeep && !canBurn {
			rows = append(rows, components.Notice("info", "info", c.Label+" can't hold subtitles and the video isn't re-encoded, so they're left out."))
		}
	}
	card := components.SectionCard("captions", "Subtitles", "Keep, convert or burn in", nil, rows...)
	if st.Subtitles.Mode == settings.SubsKeep && canKeep {
		if tracks := FileTracks(env, ffmpeg.KindSubtitle); tracks != nil {
			return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1}, card, tracks)
		}
	}
	return card
}

func boolKey(b bool) string {
	if b {
		return "+burn"
	}
	return ""
}
