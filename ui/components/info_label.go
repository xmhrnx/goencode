// Package components holds GoEncode's reusable UI pieces, one component
// per file. Stateful components are looked up by key with giode.Keyed, so
// views can build them inside the per-frame view function.
package components

import (
	"gioui.org/layout"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/tooltip"

	"github.com/xmhrnx/goencode/ui/theme"
)

// infoLabel is a setting's name with an info glyph; hovering either (or
// long-pressing on a touch screen) shows the setting's explanation for the
// chosen codec.
type infoLabel struct {
	label string
	tip   *tooltip.Tooltip
}

// InfoLabel returns the label of the setting identified by key, with help
// shown in a popover on hover. The help text may change every frame (it
// follows the chosen encoder).
func InfoLabel(key any, label, help string) giode.Element {
	l := giode.Keyed("info-label", key, func() *infoLabel {
		l := &infoLabel{}
		l.tip = giode.Tooltip(giode.Raw(l.layoutTrigger), help, giode.Styles{
			Background:   theme.TooltipBg,
			Color:        "theme.text",
			FontSize:     12.5,
			MaxWidth:     300,
			Padding:      giode.SymmetricInset(12, 9),
			BorderRadius: theme.RadiusSm,
		}).Placement(giode.PlaceAuto)
		return l
	})
	l.label = label
	if help == "" {
		return giode.Text(label, theme.Label)
	}
	l.tip.SetContent(help)
	return l.tip
}

func (l *infoLabel) layoutTrigger(gtx layout.Context) layout.Dimensions {
	iconColor := giode.Color("theme.muted/0.7")
	textColor := giode.Color("theme.text")
	if l.tip.Hovered() || l.tip.Visible() {
		iconColor, textColor = "theme.primary", "theme.primary"
	}
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Cursor: giode.CursorDefault},
		giode.Text(l.label, theme.Label.With(giode.Styles{Color: textColor})),
		giode.Icon("info", giode.Styles{Width: 13, Color: iconColor}),
	).Layout(gtx)
}
