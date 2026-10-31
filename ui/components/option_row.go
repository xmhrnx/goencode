package components

import (
	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/ui/theme"
)

// OptionRow lays out a setting: its label on the left and its control on
// the right, or stacked when the panel is narrow. A switch is small enough
// to stay on the label's line.
func OptionRow(label, control giode.Element) giode.Element {
	if _, ok := control.(*switchField); ok {
		return giode.Row(giode.Styles{Gap: theme.S3, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap, MinHeight: 38},
			giode.Box(giode.Styles{FlexGrow: 1}, label),
			giode.Box(giode.Styles{Width: 48}, control),
		)
	}
	return giode.Responsive(
		giode.Breakpoint{MaxWidth: 360, Name: "compact", View: func() giode.Element {
			return giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
				label,
				giode.Box(giode.Styles{Width: -1}, control),
			)
		}},
		giode.Breakpoint{View: func() giode.Element {
			return giode.Row(giode.Styles{Gap: theme.S3, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap, MinHeight: 38},
				giode.Box(giode.Styles{FlexGrow: 1}, label),
				giode.Box(giode.Styles{Width: 210}, control),
			)
		}},
	)
}
