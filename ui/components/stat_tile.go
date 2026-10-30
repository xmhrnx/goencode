package components

import (
	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/ui/theme"
)

// StatTile shows one number with its label: "Projected size  1.2 GB".
func StatTile(icon, label, value, hint string) giode.Element {
	rows := []giode.Element{
		giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Icon(icon, giode.Styles{Width: 13, Color: "theme.muted"}),
			giode.Text(label, theme.Micro)),
		giode.Text(value, giode.Styles{FontSize: 18, FontWeight: giode.FontWeightBold, Color: "theme.text"}),
	}
	if hint != "" {
		rows = append(rows, giode.Text(hint, theme.Caption))
	}
	return giode.Stack(giode.Styles{Gap: 4, FlexGrow: 1, MinWidth: 120, Padding: giode.UniformInset(theme.S3),
		Background: theme.Sunken, BorderRadius: theme.RadiusMd}, rows...)
}
