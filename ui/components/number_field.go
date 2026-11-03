package components

import (
	"strconv"
	"strings"

	"gioui.org/layout"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"github.com/xmhrnx/goencode/ui/theme"
)

// numberField is a text field for a number with an optional unit. Empty
// means "not set" (the encoder's default), shown through the placeholder.
type numberField struct {
	in       *input.Input
	unit     string
	shown    string
	onChange func(value string)
	min, max float64
}

// Number returns the numeric field identified by key. value is the current
// text ("" for unset); placeholder shows the default.
func Number(key any, value, placeholder, unit string, min, max float64, onChange func(string)) giode.Element {
	f := giode.Keyed("number", key, func() *numberField {
		f := &numberField{}
		f.in = giode.Input.Text(placeholder, theme.Field().With(giode.Styles{Padding: giode.SymmetricInset(10, 7)})).
			Keyboard(giode.KeyboardNumeric).
			OnChange(func(text string) { f.commit(text) })
		f.in.SetText(value)
		f.shown = value
		return f
	})
	f.unit, f.onChange, f.min, f.max = unit, onChange, min, max
	f.in.Placeholder(placeholder)
	if !f.in.Focused() && value != f.shown {
		f.in.SetText(value)
		f.shown = value
	}
	return f
}

func (f *numberField) commit(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		f.shown = ""
		f.onChange("")
		return
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return
	}
	if f.max > f.min && (v < f.min || v > f.max) {
		return
	}
	f.shown = text
	f.onChange(text)
}

func (f *numberField) Layout(gtx layout.Context) layout.Dimensions {
	if f.unit == "" {
		return f.in.Layout(gtx)
	}
	return giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1}, f.in),
		giode.Text(f.unit, theme.Caption),
	).Layout(gtx)
}
