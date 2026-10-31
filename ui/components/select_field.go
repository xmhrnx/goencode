package components

import (
	"runtime"
	"strings"

	"gioui.org/layout"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/dropdown"

	"github.com/xmhrnx/goencode/ui/theme"
)

// selectField is a dropdown whose options can change between frames: the
// codecs on offer change with the container, the encoders with the codec.
type selectField struct {
	sel      *dropdown.Select
	sig      string
	values   []string
	onChange func(string)
}

// sheet is when options open as a bottom sheet: only on phones. Desktops
// get the list below the field, even in a narrow settings panel, which
// SheetCompact would take for a phone.
var sheet = func() giode.SheetMode {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return giode.SheetAlways
	}
	return giode.SheetNever
}()

// Select returns the dropdown identified by key, showing labels for values
// with current selected. onChange receives the chosen value.
func Select(key any, labels, values []string, current string, onChange func(string)) giode.Element {
	f := giode.Keyed("select", key, func() *selectField { return &selectField{} })
	if sig := strings.Join(labels, "\x00"); sig != f.sig || f.sel == nil {
		f.sig = sig
		f.sel = giode.Input.Select(theme.Field().With(giode.Styles{Width: -1}), labels...).
			Sheet(sheet).
			OnChange(func(i int) {
				if i >= 0 && i < len(f.values) && f.onChange != nil {
					f.onChange(f.values[i])
				}
			})
	}
	f.values, f.onChange = values, onChange
	for i, v := range values {
		if v == current {
			if !f.sel.Opened() && f.sel.Selected() != i {
				f.sel.SetSelected(i)
			}
			break
		}
	}
	return f
}

func (f *selectField) Layout(gtx layout.Context) layout.Dimensions { return f.sel.Layout(gtx) }
