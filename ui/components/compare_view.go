package components

import (
	"image"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/xmhrnx/goencode/ui/theme"
)

// CompareView shows the same frame from the source and the preview encode
// side by side (stacked on narrow screens), to judge the quality.
func CompareView(before, after image.Image) giode.Element {
	frame := func(label string, img image.Image) giode.Element {
		var pic giode.Element = giode.Box(giode.Styles{Width: -1, AspectRatio: 16.0 / 9, Background: theme.Sunken, BorderRadius: theme.RadiusMd})
		if img != nil {
			b := img.Bounds()
			pic = giode.Box(giode.Styles{Width: -1, BorderRadius: theme.RadiusMd, Overflow: giode.OverflowHidden, Background: "#000"},
				giode.Image(img, giode.Styles{Width: -1, AspectRatio: properties.AspectRatio(float32(b.Dx()) / float32(max(1, b.Dy()))), Fit: giode.FitContain}))
		}
		return giode.Stack(giode.Styles{Gap: 6, FlexGrow: 1, Width: -1}, giode.Text(label, theme.Micro), pic)
	}
	return giode.Responsive(
		giode.Breakpoint{MaxWidth: 520, Name: "compact", View: func() giode.Element {
			return giode.Stack(giode.Styles{Gap: theme.S3, Width: -1}, frame("Source", before), frame("Encoded", after))
		}},
		giode.Breakpoint{View: func() giode.Element {
			return giode.Row(giode.Styles{Gap: theme.S3, Width: -1, Wrap: giode.WrapNoWrap},
				giode.Box(giode.Styles{FlexGrow: 1}, frame("Source", before)),
				giode.Box(giode.Styles{FlexGrow: 1}, frame("Encoded", after)))
		}},
	)
}
