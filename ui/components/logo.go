package components

import (
	"bytes"
	"image"
	"image/png"
	"sync"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/xmhrnx/goencode/ui/icons"
	"github.com/xmhrnx/goencode/ui/theme"
)

var (
	logoOnce sync.Once
	logoImg  image.Image
)

// Logo is the app icon with the GoEncode word mark; size is the icon's
// edge in dp, and showName adds the name beside it.
func Logo(size int, showName bool) giode.Element {
	logoOnce.Do(func() { logoImg, _ = png.Decode(bytes.NewReader(icons.Logo)) })
	mark := giode.Image(logoImg, giode.Styles{Width: properties.Width(size), Height: properties.Height(size)})
	if !showName {
		return mark
	}
	return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		mark,
		giode.Row(giode.Styles{Wrap: giode.WrapNoWrap},
			giode.Text("Go", theme.Title.With(giode.Styles{FontSize: 17, FontWeight: giode.FontWeightBold})),
			giode.Text("Encode", theme.Title.With(giode.Styles{FontSize: 17, FontWeight: giode.FontWeightBold, Color: "theme.primary"})),
		),
	)
}
