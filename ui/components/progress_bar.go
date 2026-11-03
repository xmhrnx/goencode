package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/xmhrnx/goencode/ui/theme"
)

// ProgressBar is a slim bar filled with the brand gradient.
func ProgressBar(fraction float64, height int) giode.Element {
	pct := int(fraction*100 + 0.5)
	h := properties.Height(height)
	fill := giode.Styles{Height: h, BorderRadius: theme.Pill, BackgroundGradient: theme.Gradient(90)}
	if pct <= 0 {
		fill.Display = giode.DisplayNone
	} else {
		fill.WidthPct = properties.WidthPct(min(pct, 100))
	}
	return giode.Box(giode.Styles{Width: -1, Height: h, BorderRadius: theme.Pill, Background: theme.Subtle, Overflow: giode.OverflowHidden},
		giode.Box(fill))
}
