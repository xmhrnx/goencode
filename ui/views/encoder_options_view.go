package views

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/accordion"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// EncoderOptions lists the options of the chosen video or audio encoder:
// GoEncode's curated ones first, then under "More" the rest of what ffmpeg
// reports for that encoder, each with its explanation on hover.
func EncoderOptions(env *Env, kind ffmpeg.MediaKind) giode.Element {
	s := env.Store
	p, enc, get, set := s.VideoProfile(), s.Settings.Video.Encoder, (*settings.Settings).VideoOption, (*settings.Settings).SetVideoOption
	if kind == ffmpeg.KindAudio {
		p, enc, get, set = s.AudioProfile(), s.Settings.Audio.Encoder, (*settings.Settings).AudioOption, (*settings.Settings).SetAudioOption
	}
	control := func(o catalog.Option) giode.Element {
		key := o.Key
		return components.OptionControl("opt/"+enc, o, get(&s.Settings, key), func(v string) {
			s.Update(func(x *settings.Settings) { set(x, key, v) })
		})
	}

	var simple []giode.Element
	more := 0
	for _, o := range p.Options {
		if o.Simple {
			simple = append(simple, control(o))
		} else {
			more++
		}
	}
	help := s.EncoderHelp(enc)
	extra := 0
	if help != nil {
		for _, av := range help.Options {
			if !p.IsManaged(av.Name) {
				extra++
			}
		}
	}

	rows := simple
	if len(rows) == 0 && more == 0 && extra == 0 {
		msg := p.Label + " has no options of its own."
		if help == nil {
			msg = "Reading " + p.Label + "'s options from ffmpeg…"
		}
		rows = append(rows, giode.Text(msg, theme.Caption))
	}
	if more > 0 || extra > 0 {
		acc := giode.Keyed("enc-more", enc, func() *accordion.Accordion {
			a := giode.Accordion(giode.Styles{Width: -1})
			if more > 0 {
				a.Item("More "+p.Label+" options", func() giode.Element {
					return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1, Padding: giode.SymmetricInset(0, theme.S2)}, moreRows(env, kind, false)...)
				})
			}
			a.Item("All ffmpeg options", func() giode.Element {
				return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1, Padding: giode.SymmetricInset(0, theme.S2)}, moreRows(env, kind, true)...)
			})
			return a
		})
		rows = append(rows, acc)
	}

	title, icon := "Encoder options", "tune"
	if kind == ffmpeg.KindAudio {
		title = "Audio encoder options"
	}
	return components.SectionCard(icon, title, p.Label+" · hover a name to learn what it does", nil, rows...)
}

// moreRows builds the rows inside the accordion each frame, so they show
// current values: the curated non-simple options, or the raw ffmpeg ones.
func moreRows(env *Env, kind ffmpeg.MediaKind, raw bool) []giode.Element {
	s := env.Store
	p, enc, get, set := s.VideoProfile(), s.Settings.Video.Encoder, (*settings.Settings).VideoOption, (*settings.Settings).SetVideoOption
	if kind == ffmpeg.KindAudio {
		p, enc, get, set = s.AudioProfile(), s.Settings.Audio.Encoder, (*settings.Settings).AudioOption, (*settings.Settings).SetAudioOption
	}
	var opts []catalog.Option
	if raw {
		if help := s.EncoderHelp(enc); help != nil {
			for _, av := range help.Options {
				if !p.IsManaged(av.Name) {
					opts = append(opts, catalog.FromAVOption(av))
				}
			}
		}
		if len(opts) == 0 {
			return []giode.Element{giode.Text("ffmpeg reports no further options for "+p.Label+".", theme.Caption)}
		}
	} else {
		for _, o := range p.Options {
			if !o.Simple {
				opts = append(opts, o)
			}
		}
	}
	rows := make([]giode.Element, 0, len(opts))
	for _, o := range opts {
		key := o.Key
		rows = append(rows, components.OptionControl("opt/"+enc, o, get(&s.Settings, key), func(v string) {
			s.Update(func(x *settings.Settings) { set(x, key, v) })
		}))
	}
	return rows
}
