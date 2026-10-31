package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"gioui.org/layout"

	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/jobs"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/theme"
)

// FileCardProps is what a file card shows and does.
type FileCardProps struct {
	Item     *jobs.Item
	Plan     *settings.Plan
	Selected bool
	Compact  bool
	CanOpen  bool // the result plays in another app
	CanShow  bool // the result can be shown in a file manager; otherwise it's shared
	OnSelect func()
	OnRemove func()
	OnCancel func()
	OnOpen   func()
	OnReveal func() // shows the result in the file manager, or shares it
}

type fileCard struct {
	press *giode.PressableElement
	props FileCardProps
}

// FileCard is one file in the batch: thumbnail, streams, what the recipe
// will do with it, and its progress or result.
func FileCard(p FileCardProps) giode.Element {
	c := giode.Keyed("file-card", p.Item.ID, func() *fileCard {
		c := &fileCard{}
		c.press = giode.Pressable(c.content).OnClick(func(giode.ClickEvent) {
			if c.props.OnSelect != nil {
				c.props.OnSelect()
			}
		})
		return c
	})
	c.props = p
	c.press.Label(p.Item.Name)
	return c
}

func (c *fileCard) Layout(gtx layout.Context) layout.Dimensions { return c.press.Layout(gtx) }

func (c *fileCard) content(s giode.PressState) giode.Element {
	p := c.props
	it := p.Item
	st := theme.Card().With(giode.Styles{Padding: giode.UniformInset(theme.S3), Gap: theme.S3, Width: -1, Wrap: giode.WrapNoWrap,
		Transition: giode.StyleTransition{Duration: 120e6}})
	if s.Hovered {
		st.Background = theme.Hover
	}
	if p.Selected {
		st.BorderColor, st.BorderWidth = theme.PrimaryLine, 1
		st.Background = theme.PrimarySoft
	}

	thumbW, thumbH := 124, 70
	if p.Compact {
		thumbW, thumbH = 88, 50
	}
	icon := "film"
	if it.Info != nil {
		if _, ok := it.Info.Video(); !ok {
			icon = "music"
		}
	}
	thumb := Thumbnail(it.Thumb, thumbW, thumbH, icon)
	if it.Status == jobs.Running {
		pct := fmt.Sprintf("%d%%", int(it.Progress.Fraction*100))
		thumb = giode.Box(giode.Styles{Position: giode.PositionRelative}, thumb,
			giode.Box(giode.Styles{Position: giode.PositionAbsolute, Left: giode.Flush, Top: giode.Flush, Right: giode.Flush, Bottom: giode.Flush,
				Background: "#0008", BorderRadius: theme.RadiusMd, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
				giode.Text(pct, giode.Styles{Color: "#fff", FontSize: 15, FontWeight: giode.FontWeightBold})))
	}

	actions := []giode.Element{StatusPill(it.Status)}
	if it.Status.Active() {
		actions = append(actions, Button(fmt.Sprint("cancel", it.ID), ButtonSpec{Icon: "stop", Kind: Ghost, Small: true, Tooltip: "Stop this file", OnClick: p.OnCancel}))
	} else {
		actions = append(actions, Button(fmt.Sprint("remove", it.ID), ButtonSpec{Icon: "x", Kind: Ghost, Small: true, Tooltip: "Remove from the list", OnClick: p.OnRemove}))
	}

	title := giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1},
			giode.Text(it.Name, giode.Styles{FontSize: 13.5, FontWeight: giode.FontWeightSemiBold, Color: "theme.text", TextOverflow: giode.TextOverflowEllipsis})),
		giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, actions...),
	)

	body := []giode.Element{title}
	if chips := streamChips(it.Info); len(chips) > 0 {
		body = append(body, giode.Row(giode.Styles{Gap: 6, Wrap: giode.WrapWrap}, chips...))
	}
	body = append(body, c.detail()...)

	return giode.Row(st, thumb, giode.Stack(giode.Styles{Gap: 7, FlexGrow: 1}, body...))
}

// detail is the line under the chips: what will happen, the progress, the
// result or the error.
func (c *fileCard) detail() []giode.Element {
	p := c.props
	it := p.Item
	var out []giode.Element
	meta := []string{store.FormatBytes(it.Size)}
	if it.Info != nil && it.Info.Duration > 0 {
		meta = append([]string{ffmpeg.FormatDuration(it.Info.Duration)}, meta...)
	}

	switch it.Status {
	case jobs.Running:
		pr := it.Progress
		line := fmt.Sprintf("%s / %s", ffmpeg.FormatDuration(pr.Position), ffmpeg.FormatDuration(it.Info.Duration))
		if pr.Speed > 0 {
			line += fmt.Sprintf(" · %.1f×", pr.Speed)
		}
		if pr.ETA > 0 {
			line += " · " + ffmpeg.FormatDuration(pr.ETA) + " left"
		}
		if pr.Size > 0 {
			line += " · " + store.FormatBytes(pr.Size)
		}
		out = append(out, ProgressBar(pr.Fraction, 5), giode.Text(line, theme.Caption))
	case jobs.Done:
		ratio := ""
		if r := it.Ratio(); r > 0 {
			ratio = fmt.Sprintf(" (%+.0f%%)", (r-1)*100)
		}
		res := fmt.Sprintf("→ %s · %s%s · %s", filepath.Base(it.Output), store.FormatBytes(it.OutSize), ratio, ffmpeg.FormatDuration(it.Elapsed()))
		row := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(res, theme.Caption.With(giode.Styles{Color: "theme.success", TextOverflow: giode.TextOverflowEllipsis})))}
		if p.CanOpen {
			row = append(row, Button(fmt.Sprint("open", it.ID), ButtonSpec{Icon: "play", Kind: Ghost, Small: true, Tooltip: "Play the result", OnClick: p.OnOpen}))
		}
		revealTip, revealIcon := "Show in folder", "folder"
		if !p.CanShow {
			revealTip, revealIcon = "Share or save", "share"
		}
		row = append(row, Button(fmt.Sprint("reveal", it.ID), ButtonSpec{Icon: revealIcon, Kind: Ghost, Small: true, Tooltip: revealTip, OnClick: p.OnReveal}))
		out = append(out, giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap}, row...))
	case jobs.Failed, jobs.Invalid:
		out = append(out, giode.Text(it.Err, theme.Caption.With(giode.Styles{Color: "theme.danger", MaxLines: 3})))
	default:
		if p.Plan != nil {
			if sum := p.Plan.Summary(); sum != "" {
				meta = append(meta, sum)
			}
		}
		out = append(out, giode.Text(strings.Join(meta, "  ·  "), theme.Caption.With(giode.Styles{MaxLines: 2})))
		if p.Plan != nil && it.Status != jobs.Queued {
			for _, w := range p.Plan.Warnings() {
				out = append(out, giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
					giode.Icon("alert", giode.Styles{Width: 12, Color: "theme.warning"}),
					giode.Text(w, theme.Caption.With(giode.Styles{Color: "theme.warning", MaxLines: 2}))))
			}
		}
	}
	return out
}

// streamChips summarizes a file's streams: "HEVC 2160p", "2 audio" ...
func streamChips(info *ffmpeg.MediaInfo) []giode.Element {
	if info == nil {
		return nil
	}
	var chips []giode.Element
	if v, ok := info.Video(); ok {
		label := shortCodec(v.Codec)
		if v.Height > 0 {
			label += fmt.Sprintf(" · %dp", v.Height)
		}
		if v.FrameRate > 0 {
			label += fmt.Sprintf(" · %.3g fps", v.FrameRate)
		}
		if v.HDR {
			label += " · HDR"
		} else if v.BitDepth >= 10 {
			label += " · 10-bit"
		}
		chips = append(chips, Pill(label, "primary", "video"))
	}
	if a := info.StreamsOf(ffmpeg.KindAudio); len(a) > 0 {
		label := shortCodec(a[0].Codec)
		if a[0].Channels > 0 {
			label += " " + channels(a[0].Channels)
		}
		if len(a) > 1 {
			label += fmt.Sprintf(" +%d", len(a)-1)
		}
		chips = append(chips, Pill(label, "info", "speaker"))
	}
	if s := info.StreamsOf(ffmpeg.KindSubtitle); len(s) > 0 {
		chips = append(chips, Pill(fmt.Sprintf("%d subs", len(s)), "", "captions"))
	}
	if info.Chapters > 0 {
		chips = append(chips, Pill(fmt.Sprintf("%d chapters", info.Chapters), "", "bookmark"))
	}
	return chips
}

func shortCodec(id string) string {
	l := catalog.Codec(id, "").Label
	if i := strings.Index(l, " / "); i > 0 {
		return l[:i]
	}
	if i := strings.Index(l, " ("); i > 0 {
		return l[:i]
	}
	return l
}

func channels(n int) string {
	switch n {
	case 1:
		return "mono"
	case 2:
		return "stereo"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	}
	return fmt.Sprintf("%dch", n)
}
