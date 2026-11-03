package components

import (
	"gioui.org/layout"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/theme"
)

type presetCard struct {
	press    *giode.PressableElement
	preset   settings.Preset
	selected bool
	// minHeight is set by EqualRow so the cards of a row line up.
	minHeight float32
}

// PresetCard is one recipe of the simple mode; selected shows it's the one
// in use.
func PresetCard(p settings.Preset, selected bool, onPick func()) giode.Element {
	c := giode.Keyed("preset-card", p.ID, func() *presetCard {
		c := &presetCard{}
		c.press = giode.Pressable(c.content)
		return c
	})
	c.preset, c.selected, c.minHeight = p, selected, 0
	c.press.Label(p.Name).OnClick(func(giode.ClickEvent) { onPick() })
	return c
}

func (c *presetCard) Layout(gtx layout.Context) layout.Dimensions { return c.press.Layout(gtx) }

func (c *presetCard) SetMinHeight(dp float32) { c.minHeight = dp }

func (c *presetCard) content(s giode.PressState) giode.Element {
	p := c.preset
	st := giode.Styles{Width: -1, Padding: giode.UniformInset(theme.S3), Gap: theme.S2, BorderRadius: theme.RadiusMd,
		Background: theme.Raised, BorderColor: "theme.border", BorderWidth: 1, Cursor: giode.CursorPointer}
	if c.minHeight > 0 {
		st.Height = properties.Height(c.minHeight)
	}
	iconBg, iconFg := giode.Color(theme.PrimarySoft), giode.Color("theme.primary")
	if s.Hovered {
		st.Background, st.BorderColor = theme.Hover, theme.PrimaryLine
	}
	if c.selected {
		st.BorderColor, st.BorderWidth, st.Background = "theme.primary", 2, theme.PrimarySoft
		st.BoxShadow = giode.BoxShadow{Y: 6, Blur: 20, Color: "#7C5CFF33"}
		iconFg = "#FFFFFF"
	}
	icon := giode.Box(giode.Styles{Width: 32, Height: 32, BorderRadius: 10, Background: iconBg, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
		giode.Icon(p.Icon, giode.Styles{Width: 17, Color: iconFg}))
	if c.selected {
		icon = giode.Box(giode.Styles{Width: 32, Height: 32, BorderRadius: 10, BackgroundGradient: theme.Gradient(135), Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			giode.Icon(p.Icon, giode.Styles{Width: 17, Color: iconFg}))
	}
	tagTone := ""
	if c.selected {
		tagTone = "primary"
	}
	return giode.Stack(st,
		giode.Row(giode.Styles{Justify: giode.JustifySpaceBetween, Align: giode.AlignStart, Width: -1, Wrap: giode.WrapNoWrap},
			icon, Pill(p.Tag, tagTone, "")),
		giode.Text(p.Name, theme.Heading),
		giode.Text(p.Description, theme.Caption.With(giode.Styles{MaxLines: 3, LineHeight: 16})),
	)
}
