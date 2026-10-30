package components

import (
	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/ui/theme"
)

// Pill is a small rounded label: a stream's codec, a status, a tag. tone
// is a theme status ("success", "warning", "danger", "info", "primary") or
// "" for neutral; icon may be empty.
func Pill(text, tone, icon string) giode.Element {
	bg, fg := theme.Tone(tone)
	children := []giode.Element{}
	if icon != "" {
		children = append(children, giode.Icon(icon, giode.Styles{Width: 12, Color: fg}))
	}
	children = append(children, giode.Text(text, giode.Styles{FontSize: 11.5, FontWeight: giode.FontWeightMedium, Color: fg, MaxLines: 1}))
	return giode.Row(giode.Styles{Gap: 5, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Background: bg, BorderRadius: theme.Pill,
		Padding: giode.SymmetricInset(9, 3)}, children...)
}
