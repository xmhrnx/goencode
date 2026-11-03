package views

import (
	"fmt"
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// PrefsPanel holds the app's preferences: which ffmpeg to use, how many
// files encode at once, the theme and notifications.
type PrefsPanel struct {
	env   *Env
	path  string // ffmpeg path being typed, applied with the button
	dirty bool
}

// NewPrefsPanel creates the preferences panel.
func NewPrefsPanel(env *Env) *PrefsPanel {
	return &PrefsPanel{env: env}
}

// View builds the panel's content.
func (p *PrefsPanel) View() giode.Element {
	s := p.env.Store
	if !p.dirty {
		p.path = s.Prefs.FFmpegPath
	}

	engine := []giode.Element{}
	switch s.Engine {
	case store.EngineReady:
		hw := "testing…"
		if s.HWReady {
			hw = fmt.Sprint(len(s.Caps.WorkingHardware()), " GPU encoder families")
		}
		engine = append(engine,
			components.Notice("success", "check", "ffmpeg "+s.Bins.Version+" · "+strconv.Itoa(len(s.Caps.Codecs))+" codecs · "+hw),
			giode.Text(s.Bins.FFmpeg, giode.Styles{FontSize: 11, Color: "theme.muted", MaxLines: 2}))
	default:
		engine = append(engine, components.EngineBanner(s.Engine, nil, s.LoadEngine))
		if s.EngineErr != "" {
			engine = append(engine, giode.Text(s.EngineErr, theme.Caption))
		}
	}
	if p.env.Desktop() {
		engine = append(engine,
			components.OptionRow(components.InfoLabel("pref-ffmpeg", "ffmpeg location",
				"Path to the ffmpeg program, or the folder holding ffmpeg and ffprobe. Empty: GoEncode uses the one bundled next to it, or the one on your PATH."),
				components.TextInput("pref-ffmpeg", p.path, "Automatic", func(v string) { p.path, p.dirty = v, true })),
			giode.Row(giode.Styles{Gap: theme.S2},
				components.Button("pref-apply", components.ButtonSpec{Label: "Use this ffmpeg", Icon: "check", Kind: components.Secondary, Small: true,
					Disabled: !p.dirty, OnClick: func() {
						next := s.Prefs
						next.FFmpegPath = p.path
						p.dirty = false
						s.SetPrefs(next)
					}}),
				components.Button("pref-retry", components.ButtonSpec{Label: "Detect again", Icon: "retry", Kind: components.Ghost, Small: true,
					Disabled: s.Running, OnClick: s.LoadEngine})))
	}

	set := func(fn func(*platform.Prefs)) {
		next := s.Prefs
		fn(&next)
		s.SetPrefs(next)
	}
	return giode.Stack(giode.Styles{Gap: theme.S4, Width: -1},
		components.SectionCard("chip", "Engine", "The ffmpeg GoEncode drives", nil, engine...),
		components.SectionCard("gauge", "Performance", "", nil,
			components.OptionRow(components.InfoLabel("pref-jobs", "Files at once",
				"How many files encode in parallel. Most encoders already use every CPU core, so 1 is usually fastest; raise it for many small files or when a GPU encoder leaves the CPU idle."),
				giode.Box(giode.Styles{Width: -1}, components.Segmented("pref-jobs", []string{"1", "2", "3", "4"}, []string{"1", "2", "3", "4"}, nil,
					strconv.Itoa(max(1, s.Prefs.Concurrency)), func(v string) { set(func(x *platform.Prefs) { x.Concurrency, _ = strconv.Atoi(v) }) })))),
		components.SectionCard("sun", "Appearance & alerts", "", nil,
			components.OptionRow(giode.Text("Theme", theme.Label),
				giode.Box(giode.Styles{Width: -1}, components.Segmented("pref-theme", []string{"System", "Dark", "Light"}, []string{"system", "dark", "light"}, nil,
					themeOrSystem(s.Prefs.Theme), func(v string) {
						set(func(x *platform.Prefs) { x.Theme = v })
						theme.Apply(v)
					}))),
			components.OptionRow(components.InfoLabel("pref-notify", "Notify when done", "Shows a system notification when a batch finishes, so you can leave GoEncode in the background."),
				components.Switch("pref-notify", s.Prefs.Notify, func(on bool) { set(func(x *platform.Prefs) { x.Notify = on }) }))),
	)
}

func themeOrSystem(t string) string {
	if t == "" {
		return "system"
	}
	return t
}
