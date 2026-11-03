package components

import (
	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/ui/theme"
)

// DropZoneProps configures the empty state of the file list.
type DropZoneProps struct {
	Dragging    bool // files are being dragged over the window
	CanDrop     bool // the platform supports drag and drop
	CanAddDir   bool
	OnAdd       func()
	OnAddFolder func()
}

// DropZone is the inviting empty state: drop files here, or pick them.
func DropZone(p DropZoneProps) giode.Element {
	border := giode.Color(theme.Subtle)
	bg := giode.Color("transparent")
	if p.Dragging {
		border, bg = "theme.primary", theme.PrimarySoft
	}
	headline := "Add media to get started"
	sub := "Video and audio in any format ffmpeg reads."
	if p.CanDrop {
		headline = "Drop video or audio files here"
		sub = "Or folders — every media file inside is added. Any format ffmpeg reads."
	}
	buttons := []giode.Element{
		Button("dz-add", ButtonSpec{Label: "Choose files", Icon: "plus", Kind: Primary, OnClick: p.OnAdd}),
	}
	if p.CanAddDir {
		buttons = append(buttons, Button("dz-folder", ButtonSpec{Label: "Add folder", Icon: "folder", Kind: Secondary, OnClick: p.OnAddFolder}))
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(theme.S6)},
		giode.Stack(giode.Styles{
			Width: -1, Height: -1, MinHeight: 280, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: theme.S4,
			BorderColor: border, BorderWidth: 2, BorderRadius: theme.RadiusXl, Background: bg, Padding: giode.UniformInset(theme.S8),
		},
			giode.Box(giode.Styles{Width: 76, Height: 76, BorderRadius: 24, BackgroundGradient: theme.Gradient(135),
				Align: giode.AlignCenter, Justify: giode.JustifyCenter, BoxShadow: giode.BoxShadow{Y: 10, Blur: 30, Color: "#7C5CFF66"}},
				giode.Icon("upload", giode.Styles{Width: 34, Color: "#FFFFFF"})),
			giode.Text(headline, theme.Title.With(giode.Styles{FontSize: 18, TextAlign: giode.TextAlignCenter})),
			giode.Text(sub, theme.Caption.With(giode.Styles{TextAlign: giode.TextAlignCenter, MaxWidth: 380})),
			giode.Row(giode.Styles{Gap: theme.S3, Justify: giode.JustifyCenter}, buttons...),
		))
}
