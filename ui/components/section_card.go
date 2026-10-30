package components

import (
	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/ui/theme"
)

// SectionCard groups related settings under a title with an icon, and an
// optional element on the right of the title (a switch, a badge).
func SectionCard(icon, title, subtitle string, trailing giode.Element, children ...giode.Element) giode.Element {
	head := []giode.Element{
		giode.Box(giode.Styles{Width: 30, Height: 30, BorderRadius: theme.RadiusSm, Background: theme.PrimarySoft,
			Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			giode.Icon(icon, giode.Styles{Width: 16, Color: "theme.primary"})),
	}
	titles := []giode.Element{giode.Text(title, theme.Heading)}
	if subtitle != "" {
		titles = append(titles, giode.Text(subtitle, theme.Caption))
	}
	head = append(head, giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1}, titles...))
	if trailing != nil {
		head = append(head, trailing)
	}
	body := []giode.Element{giode.Row(giode.Styles{Gap: theme.S3, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap}, head...)}
	if len(children) > 0 {
		body = append(body, giode.Divider(giode.Styles{Color: theme.Subtle}))
		body = append(body, children...)
	}
	return giode.Stack(theme.Card().With(giode.Styles{Gap: theme.S3, Width: -1}), body...)
}
