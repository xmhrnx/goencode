package views

import (
	"runtime"

	"gioui.org/layout"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/titlebar"

	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// TitleBar is GoEncode's own window frame on desktops. The brand, the
// simple/advanced switch and the header actions sit in the draggable
// caption next to the window buttons (traffic lights on the left on
// macOS). Phones have no window frame and use Header instead.
type TitleBar struct {
	env     *Env
	bar     *titlebar.TitleBar
	compact bool
}

// NewTitleBar creates the title bar of the app's frameless window.
func NewTitleBar(env *Env) *TitleBar {
	t := &TitleBar{env: env}
	variant := giode.TitleBarWindows
	if runtime.GOOS == "darwin" {
		variant = giode.TitleBarMac
	}
	t.bar = giode.TitleBar(env.App, "", giode.Styles{Height: 52, Background: "theme.background", BorderColor: "theme.border", BorderWidth: 1}).
		Variant(variant).
		Icon(giode.Box(giode.Styles{Padding: giode.Inset{Left: theme.S2, Right: theme.S2}}, components.Logo(26, true))).
		Trailing(dynamic(t.trailing))
	return t
}

// Compact switches to the actions of a narrow window.
func (t *TitleBar) Compact(on bool) *TitleBar {
	t.compact = on
	return t
}

func (t *TitleBar) Layout(gtx layout.Context) layout.Dimensions { return t.bar.Layout(gtx) }

func (t *TitleBar) trailing() giode.Element {
	items := []giode.Element{}
	if !t.compact {
		items = append(items, giode.Box(giode.Styles{Width: 230}, ModeSwitch(t.env, "titlebar")))
	}
	items = append(items, headerActions(t.env, t.compact)...)
	return giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Right: theme.S3}}, items...)
}

// dynamicEl rebuilds its content every frame, for the slots of stateful
// components that are filled once.
type dynamicEl func() giode.Element

func dynamic(fn func() giode.Element) giode.Element { return dynamicEl(fn) }

func (d dynamicEl) Layout(gtx layout.Context) layout.Dimensions { return d().Layout(gtx) }
