package views

import (
	"fmt"

	"gioui.org/layout"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/jobs"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// FileList is the batch: a header with totals and the file cards, or the
// drop zone while it's empty.
type FileList struct {
	env     *Env
	scroll  *scroll.Scroll
	compact bool
}

// NewFileList creates the file list.
func NewFileList(env *Env) *FileList {
	l := &FileList{env: env}
	l.scroll = giode.Scroll(l.cards, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, ThumbColor: theme.Subtle, AutoHide: true}})
	return l
}

// Compact switches to the phone layout of the cards.
func (l *FileList) Compact(on bool) *FileList {
	l.compact = on
	return l
}

func (l *FileList) Layout(gtx layout.Context) layout.Dimensions {
	s := l.env.Store
	if len(s.Items) == 0 {
		_, dragging := l.env.App.FileDrag()
		return components.DropZone(components.DropZoneProps{
			Dragging: dragging, CanDrop: l.env.Desktop(), CanAddDir: l.env.Desktop(),
			OnAdd: l.env.AddFiles, OnAddFolder: l.env.AddFolder,
		}).Layout(gtx)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1},
		l.header(),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, l.scroll),
	).Layout(gtx)
}

func (l *FileList) header() giode.Element {
	s := l.env.Store
	count, size, dur := s.Totals()
	summary := fmt.Sprintf("%d file", count)
	if count != 1 {
		summary += "s"
	}
	summary += " · " + store.FormatBytes(size)
	if dur > 0 {
		summary += " · " + ffmpeg.FormatDuration(dur)
	}
	finished := 0
	for _, it := range s.Items {
		if it.Status == jobs.Done || it.Status == jobs.Failed || it.Status == jobs.Canceled || it.Status == jobs.Invalid {
			finished++
		}
	}
	actions := []giode.Element{}
	if finished > 0 {
		actions = append(actions, components.Button("clear-finished", components.ButtonSpec{Label: "Clear finished", Icon: "check", Kind: components.Ghost, Small: true,
			OnClick: s.ClearFinished}))
	}
	actions = append(actions, components.Button("clear-all", components.ButtonSpec{Icon: "trash", Kind: components.Ghost, Small: true,
		Tooltip: "Remove all files", Disabled: s.Running, OnClick: s.ClearAll}))

	pad := giode.Inset{Top: theme.S2, Right: theme.S4, Bottom: theme.S2, Left: theme.S5}
	if l.compact {
		pad = giode.SymmetricInset(theme.S4, theme.S2)
	}
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: theme.S2, Wrap: giode.WrapNoWrap, Padding: pad},
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text("Files", theme.Title),
			giode.Text(summary, theme.Caption)),
		giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, actions...),
	)
}

func (l *FileList) cards() giode.Element {
	s := l.env.Store
	sel := s.SelectedItem()
	cards := make([]giode.Element, 0, len(s.Items))
	for _, it := range s.Items {
		it := it
		cards = append(cards, components.FileCard(components.FileCardProps{
			Item:     it,
			Plan:     s.Plan(it),
			Selected: sel != nil && sel.ID == it.ID && len(s.Items) > 1,
			Compact:  l.compact,
			CanOpen:  platform.CanOpen(it.Output),
			CanShow:  l.env.Desktop(),
			OnSelect: func() { s.Selected = it.ID },
			OnRemove: func() { giode.AfterLayout(func() { s.Remove(it.ID) }) },
			OnCancel: func() { s.Cancel(it.ID) },
			OnOpen:   func() { l.env.Open(it.Output) },
			OnReveal: func() { l.env.Reveal(it.Output) },
		}))
	}
	pad := giode.Inset{Top: theme.S1, Right: theme.S4, Bottom: theme.S5, Left: theme.S5}
	if l.compact {
		pad = giode.Inset{Top: theme.S1, Right: theme.S3, Bottom: theme.S4, Left: theme.S3}
	}
	return giode.Stack(giode.Styles{Gap: theme.S2 + 2, Width: -1, Padding: pad}, cards...)
}
