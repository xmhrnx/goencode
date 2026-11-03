package components

import (
	"gioui.org/layout"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toggle"
)

type switchField struct {
	sw       *toggle.Switch
	onChange func(bool)
}

// Switch returns the on/off switch identified by key.
func Switch(key any, on bool, onChange func(bool)) giode.Element {
	f := giode.Keyed("switch", key, func() *switchField {
		f := &switchField{}
		f.sw = giode.Input.Switch("", giode.Styles{Background: "theme.primary"}).OnChange(func(v bool) { f.onChange(v) })
		return f
	})
	f.onChange = onChange
	if f.sw.Checked() != on {
		f.sw.SetChecked(on)
	}
	return f
}

func (f *switchField) Layout(gtx layout.Context) layout.Dimensions {
	return giode.Row(giode.Styles{Justify: giode.JustifyEnd, Width: -1}, f.sw).Layout(gtx)
}
