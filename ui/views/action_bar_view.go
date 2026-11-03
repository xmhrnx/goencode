package views

import (
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/internal/jobs"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// ActionBar sits under the settings: the batch's progress, which GPU
// encoders work, and the Preview and Start buttons.
func ActionBar(env *Env, compact bool) giode.Element {
	s := env.Store
	fraction, running, queued := s.Progress()

	var status giode.Element
	switch {
	case s.Running:
		eta := ""
		for _, it := range s.Items {
			if it.Status == jobs.Running && it.Progress.Speed > 0 {
				eta = fmt.Sprintf(" · %.1f×", it.Progress.Speed)
				break
			}
		}
		status = giode.Stack(giode.Styles{Gap: 6, Width: -1},
			giode.Row(giode.Styles{Width: -1, Wrap: giode.WrapNoWrap},
				giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(fmt.Sprintf("Encoding %d · %d waiting%s", running, queued, eta), theme.Label)),
				giode.Text(fmt.Sprintf("%.0f%%", fraction*100), giode.Styles{FontSize: 13, FontWeight: giode.FontWeightBold, Color: "theme.primary"})),
			components.ProgressBar(fraction, 6))
	case s.Ready():
		status = hardwareInfo(env)
	default:
		status = giode.Text("ffmpeg isn't ready", theme.Caption)
	}

	hasFiles := false
	for _, it := range s.Items {
		if it.Info != nil {
			hasFiles = true
			break
		}
	}
	preview := components.Button("preview", components.ButtonSpec{Label: "Preview", Icon: "eye", Kind: components.Secondary,
		Tooltip:  "Encode a few seconds with these settings to check quality, speed and size",
		Disabled: !hasFiles || !s.Ready(), OnClick: env.OpenPreview})
	var main giode.Element
	if s.Running {
		main = components.Button("stop", components.ButtonSpec{Label: "Stop", Icon: "stop", Kind: components.Danger, Grow: true, OnClick: s.Stop})
	} else {
		label := "Start"
		if n := len(s.Items); n > 1 {
			label = fmt.Sprintf("Encode %d files", n)
		} else if n == 1 {
			label = "Encode"
		}
		main = components.Button("start", components.ButtonSpec{Label: label, Icon: "play", Kind: components.Primary, Grow: true,
			Disabled: !hasFiles || !s.Ready(), OnClick: s.Start})
	}

	pad := giode.Inset{Top: theme.S3, Right: theme.S5, Bottom: theme.S4, Left: theme.S4}
	if compact {
		pad = giode.Inset{Top: theme.S3, Right: theme.S3, Bottom: theme.S3, Left: theme.S3}
	}
	return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1, Padding: pad, Background: theme.Raised, BorderColor: "theme.border", BorderWidth: 1},
		status,
		giode.Row(giode.Styles{Gap: theme.S2, Width: -1, Wrap: giode.WrapNoWrap, Align: giode.AlignCenter},
			preview, giode.Box(giode.Styles{FlexGrow: 1}, main)),
	)
}

// hardwareInfo names the GPU encoders that passed their test.
func hardwareInfo(env *Env) giode.Element {
	s := env.Store
	if !s.HWReady {
		return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 12, Color: "theme.muted"}),
			giode.Text("Checking hardware acceleration…", theme.Caption))
	}
	hw := s.Caps.WorkingHardware()
	if len(hw) == 0 {
		return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Icon("chip", giode.Styles{Width: 13, Color: "theme.muted"}),
			giode.Text("No GPU encoder available · encoding on the CPU", theme.Caption))
	}
	return giode.Tooltip(
		giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Icon("chip", giode.Styles{Width: 13, Color: "theme.success"}),
			giode.Text("GPU ready · "+strings.Join(hw, ", "), giode.Styles{FontSize: 12, Color: "theme.text", MaxLines: 1})),
		"These hardware encoders passed a test encode on this machine. Pick one as the encoder (marked GPU) or use the GPU preset; ffmpeg "+s.Bins.Version+".",
		giode.Styles{MaxWidth: 280})
}
