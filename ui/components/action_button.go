package components

import (
	"gioui.org/layout"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/tooltip"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/xmhrnx/goencode/ui/theme"
)

// ButtonKind is a button's visual weight.
type ButtonKind int

const (
	// Primary is the brand-gradient call to action.
	Primary ButtonKind = iota
	// Secondary is a raised neutral button.
	Secondary
	// Ghost is transparent until hovered.
	Ghost
	// Danger is a destructive action.
	Danger
)

// ButtonSpec describes a button for one frame.
type ButtonSpec struct {
	Label    string
	Icon     string
	Kind     ButtonKind
	Tooltip  string // shown on hover; icon-only buttons should set it
	Disabled bool
	Small    bool
	Grow     bool // fill the available width
	OnClick  func()
}

type actionButton struct {
	press *giode.PressableElement
	tip   *tooltip.Tooltip
	spec  ButtonSpec
}

// Button returns the button identified by key.
func Button(key any, spec ButtonSpec) giode.Element {
	b := giode.Keyed("button", key, func() *actionButton {
		b := &actionButton{}
		b.press = giode.Pressable(b.content).OnClick(func(giode.ClickEvent) {
			if b.spec.OnClick != nil && !b.spec.Disabled {
				b.spec.OnClick()
			}
		})
		b.tip = giode.Tooltip(b.press, "", giode.Styles{Background: theme.TooltipBg, Color: "theme.text", FontSize: 12, MaxWidth: 260,
			Padding: giode.SymmetricInset(10, 6), BorderRadius: theme.RadiusSm})
		return b
	})
	b.spec = spec
	b.press.Disabled(spec.Disabled)
	label := spec.Label
	if label == "" {
		label = spec.Tooltip
	}
	b.press.Label(label)
	b.tip.SetContent(spec.Tooltip)
	return b
}

func (b *actionButton) Layout(gtx layout.Context) layout.Dimensions {
	if b.spec.Tooltip == "" {
		return b.press.Layout(gtx)
	}
	return b.tip.Layout(gtx)
}

func (b *actionButton) content(s giode.PressState) giode.Element {
	sp := b.spec
	pad := giode.SymmetricInset(16, 10)
	size := 15
	font := giode.Styles{FontSize: 13.5, FontWeight: giode.FontWeightSemiBold}
	if sp.Small {
		pad, size, font.FontSize = giode.SymmetricInset(11, 7), 14, 12.5
	}
	if sp.Label == "" {
		pad = giode.UniformInset(8)
		if sp.Small {
			pad = giode.UniformInset(6)
		}
	}
	st := giode.Styles{Padding: pad, BorderRadius: theme.RadiusMd, Gap: theme.S2, Align: giode.AlignCenter,
		Justify: giode.JustifyCenter, Wrap: giode.WrapNoWrap, Cursor: giode.CursorPointer}
	if sp.Grow {
		st.Width = -1
	}
	var fg giode.Color
	switch sp.Kind {
	case Primary:
		fg = "#FFFFFF"
		st.BackgroundGradient = theme.Gradient(100)
		st.BoxShadow = giode.BoxShadow{Y: 6, Blur: 18, Color: "#7C5CFF55"}
		if s.Hovered {
			st.BoxShadow = giode.BoxShadow{Y: 8, Blur: 24, Color: "#7C5CFF88"}
		}
		if s.Pressed {
			st.Opacity = 0.85
		}
	case Secondary:
		fg = "theme.text"
		st.Background, st.BorderColor, st.BorderWidth = theme.Raised, "theme.border", 1
		if s.Hovered {
			st.Background, st.BorderColor = theme.Hover, theme.PrimaryLine
		}
	case Ghost:
		fg = "theme.muted"
		if s.Hovered {
			st.Background, fg = theme.Hover, "theme.text"
		}
	case Danger:
		bg, strong := theme.Tone("danger")
		fg, st.Background = strong, bg
		if s.Hovered {
			st.Background, fg = strong, "#FFFFFF"
		}
	}
	if sp.Disabled {
		st.Opacity, st.Cursor, st.BoxShadow = 0.45, giode.CursorNotAllowed, giode.BoxShadow{}
	}
	font.Color = fg
	var kids []giode.Element
	if sp.Icon != "" {
		kids = append(kids, giode.Icon(sp.Icon, giode.Styles{Width: properties.Width(size), Color: fg}))
	}
	if sp.Label != "" {
		kids = append(kids, giode.Text(sp.Label, font))
	}
	return giode.Row(st, kids...)
}
