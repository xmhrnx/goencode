package components

import (
	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/ui/theme"
)

// Notice is an inline message box: tone "info", "warning", "danger" or
// "success".
func Notice(tone, icon, text string) giode.Element {
	bg, fg := theme.Tone(tone)
	return giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignStart, Width: -1, Wrap: giode.WrapNoWrap, Background: bg,
		BorderRadius: theme.RadiusSm, Padding: giode.SymmetricInset(12, 9)},
		giode.Icon(icon, giode.Styles{Width: 15, Color: fg}),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(text, giode.Styles{FontSize: 12.5, Color: "theme.text", LineHeight: 17})),
	)
}
