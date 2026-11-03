package components

import (
	"gioui.org/layout"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"github.com/xmhrnx/goencode/ui/theme"
)

type textField struct {
	in       *input.Input
	shown    string
	onChange func(string)
}

// TextInput returns the single-line text field identified by key. Changes
// are reported as they're typed.
func TextInput(key any, value, placeholder string, onChange func(string)) giode.Element {
	f := giode.Keyed("text", key, func() *textField {
		f := &textField{}
		f.in = giode.Input.Text(placeholder, theme.Field().With(giode.Styles{Padding: giode.SymmetricInset(10, 7)})).
			OnChange(func(t string) {
				f.shown = t
				f.onChange(t)
			})
		f.in.SetText(value)
		f.shown = value
		return f
	})
	f.onChange = onChange
	f.in.Placeholder(placeholder)
	if !f.in.Focused() && value != f.shown {
		f.in.SetText(value)
		f.shown = value
	}
	return f
}

func (f *textField) Layout(gtx layout.Context) layout.Dimensions { return f.in.Layout(gtx) }
