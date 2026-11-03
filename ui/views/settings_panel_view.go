package views

import (
	"gioui.org/layout"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// SettingsPanel is the recipe applied to every file: the format on top,
// then the simple presets or the advanced sections.
type SettingsPanel struct {
	env     *Env
	scroll  *scroll.Scroll
	section string
	compact bool
}

// NewSettingsPanel creates the settings panel.
func NewSettingsPanel(env *Env) *SettingsPanel {
	p := &SettingsPanel{env: env, section: "video"}
	p.scroll = giode.Scroll(p.content, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, ThumbColor: theme.Subtle, AutoHide: true}})
	return p
}

// Compact switches to the phone layout, which shows the simple/advanced
// switch inside the panel instead of the header.
func (p *SettingsPanel) Compact(on bool) *SettingsPanel {
	p.compact = on
	return p
}

func (p *SettingsPanel) Layout(gtx layout.Context) layout.Dimensions {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, p.scroll).Layout(gtx)
}

func (p *SettingsPanel) content() giode.Element {
	env := p.env
	s := env.Store
	pad := giode.Inset{Top: theme.S4, Right: theme.S5, Bottom: theme.S6, Left: theme.S4}
	if p.compact {
		pad = giode.Inset{Top: theme.S2, Right: theme.S3, Bottom: theme.S5, Left: theme.S3}
	}
	if !s.Ready() {
		return giode.Stack(giode.Styles{Width: -1, Padding: pad},
			components.EngineBanner(s.Engine, env.OpenPrefs, s.LoadEngine))
	}

	children := []giode.Element{}
	if p.compact {
		children = append(children, ModeSwitch(env, "panel"))
	}
	children = append(children, ContainerPicker(env))
	if !s.Prefs.Advanced {
		children = append(children, SimpleSettings(env))
	} else {
		children = append(children,
			components.Segmented("section", []string{"Video", "Audio", "Subtitles", "Output"},
				[]string{"video", "audio", "subs", "output"}, nil, p.section,
				func(v string) { p.section = v }))
		switch p.section {
		case "audio":
			children = append(children, AudioSettings(env))
		case "subs":
			children = append(children, SubtitleSettings(env), ExtrasSettings(env))
		case "output":
			children = append(children, OutputSettings(env))
		default:
			children = append(children, VideoSettings(env))
		}
	}
	return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1, Padding: pad}, children...)
}
