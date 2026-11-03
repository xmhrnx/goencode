package components

import (
	"image"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/xmhrnx/goencode/ui/theme"
)

// Thumbnail shows a video frame cropped to w×h, or, for audio files and
// before the frame is ready, a gradient tile with an icon.
func Thumbnail(img image.Image, w, h int, fallbackIcon string) giode.Element {
	W, H := properties.Width(w), properties.Height(h)
	box := giode.Styles{Width: W, Height: H, BorderRadius: theme.RadiusMd, Overflow: giode.OverflowHidden,
		Align: giode.AlignCenter, Justify: giode.JustifyCenter}
	if img != nil {
		return giode.Box(box.With(giode.Styles{Background: "#000"}),
			giode.Image(img, giode.Styles{Width: W, Height: H, Fit: giode.FitCover}))
	}
	box.BackgroundGradient = giode.Linear(135, string(theme.PrimarySoft), string(theme.Sunken))
	return giode.Box(box, giode.Icon(fallbackIcon, giode.Styles{Width: properties.Width(h * 2 / 5), Color: "theme.primary"}))
}
