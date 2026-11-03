package views

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// PreviewPanel encodes a short part of one file with the current settings
// and shows the result next to the source, with its speed and the size the
// whole file would have.
type PreviewPanel struct {
	env    *Env
	at     float64 // where the clip starts, 0..1 of the file
	length int     // seconds
}

// NewPreviewPanel creates the preview panel.
func NewPreviewPanel(env *Env) *PreviewPanel {
	return &PreviewPanel{env: env, at: 0.3, length: 10}
}

// View builds the panel's content.
func (p *PreviewPanel) View() giode.Element {
	s := p.env.Store
	it := s.SelectedItem()
	if it == nil || it.Info == nil {
		for _, x := range s.Items {
			if x.Info != nil {
				it = x
				break
			}
		}
	}
	if it == nil {
		return giode.Text("Add a file to preview.", theme.Body)
	}
	pv := &s.Preview
	dur := it.Info.Duration
	start := time.Duration(p.at * float64(dur)).Truncate(time.Second)

	labels, values := []string{}, []string{}
	for _, x := range s.Items {
		if x.Info != nil {
			labels, values = append(labels, x.Name), append(values, strconv.Itoa(x.ID))
		}
	}
	running := pv.State == store.PreviewRunning
	rows := []giode.Element{
		giode.Text("Encodes a few seconds with the current settings, so you can judge quality, speed and size before the whole batch.", theme.Caption),
	}
	if len(values) > 1 {
		rows = append(rows, components.OptionRow(giode.Text("File", theme.Label),
			components.Select("pv-file", labels, values, strconv.Itoa(it.ID), func(v string) { s.Selected, _ = strconv.Atoi(v) })))
	}
	rows = append(rows,
		giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
			components.InfoLabel("pv-at", "Start at", "Where in the file the clip begins. Pick a busy, detailed scene: it's where encoders struggle first."),
			components.Slider("pv-at", p.at, components.SliderSpec{Min: 0, Max: 1, Step: 0.001,
				Format:    func(v float64) string { return ffmpeg.FormatDuration(time.Duration(v * float64(dur))) },
				LeftLabel: "0:00", RightLabel: ffmpeg.FormatDuration(dur)},
				func(v float64) { p.at = v })),
		components.OptionRow(components.InfoLabel("pv-len", "Length", "How many seconds to encode. Longer clips predict the full size more accurately."),
			giode.Box(giode.Styles{Width: -1}, components.Segmented("pv-len", []string{"5 s", "10 s", "30 s", "60 s"}, []string{"5", "10", "30", "60"}, nil,
				strconv.Itoa(p.length), func(v string) { p.length, _ = strconv.Atoi(v) }))),
	)

	if running {
		rows = append(rows, giode.Stack(giode.Styles{Gap: 6, Width: -1},
			giode.Row(giode.Styles{Width: -1, Wrap: giode.WrapNoWrap},
				giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Encoding preview…", theme.Label)),
				giode.Text(fmt.Sprintf("%.0f%%", pv.Progress.Fraction*100), theme.Label)),
			components.ProgressBar(pv.Progress.Fraction, 6)),
			components.Button("pv-cancel", components.ButtonSpec{Label: "Cancel", Icon: "x", Kind: components.Secondary, OnClick: s.CancelPreview}))
	} else {
		label := "Encode preview"
		if pv.State == store.PreviewDone && pv.ItemID == it.ID {
			label = "Encode again"
		}
		id := it.ID
		rows = append(rows, components.Button("pv-run", components.ButtonSpec{Label: label, Icon: "play", Kind: components.Primary,
			OnClick: func() { s.StartPreview(id, start, time.Duration(p.length)*time.Second) }}))
	}

	if pv.ItemID == it.ID {
		switch pv.State {
		case store.PreviewFailed:
			rows = append(rows, components.Notice("danger", "alert", pv.Err))
		case store.PreviewDone:
			rows = append(rows, p.results(it.Info.Size)...)
		}
	}
	return giode.Stack(giode.Styles{Gap: theme.S4, Width: -1}, rows...)
}

func (p *PreviewPanel) results(sourceSize int64) []giode.Element {
	env := p.env
	pv := &env.Store.Preview
	speed := "—"
	if pv.Took > 0 {
		speed = fmt.Sprintf("%.1f×", pv.Length.Seconds()/pv.Took.Seconds())
	}
	ratio := ""
	if sourceSize > 0 && pv.Projected > 0 {
		ratio = fmt.Sprintf("%.0f%% of the source", float64(pv.Projected)/float64(sourceSize)*100)
	}
	out := pv.Output
	actions := []giode.Element{}
	if platform.CanOpen(out) {
		actions = append(actions, components.Button("pv-open", components.ButtonSpec{Label: "Play clip", Icon: "play", Kind: components.Secondary, Small: true,
			OnClick: func() { env.Open(out) }}))
	}
	shareLabel := "Show in folder"
	if !env.Desktop() {
		shareLabel = "Share clip"
	}
	actions = append(actions, components.Button("pv-share", components.ButtonSpec{Label: shareLabel, Icon: "share", Kind: components.Ghost, Small: true,
		OnClick: func() { env.Reveal(out) }}))

	rows := []giode.Element{
		giode.Row(giode.Styles{Gap: theme.S2, Width: -1},
			components.StatTile("gauge", "Speed", speed, "of real time"),
			components.StatTile("file", "Clip size", store.FormatBytes(pv.Size), ffmpeg.FormatDuration(pv.Length)),
			components.StatTile("layers", "Full file ≈", store.FormatBytes(pv.Projected), ratio)),
	}
	if pv.Before != nil || pv.After != nil {
		rows = append(rows, components.CompareView(pv.Before, pv.After))
	}
	rows = append(rows, giode.Row(giode.Styles{Gap: theme.S2}, actions...))
	return rows
}
