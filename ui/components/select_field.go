package components

import (
	"runtime"
	"strings"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/dropdown"
	"github.com/mehran9675/giode/components/kit"
	"github.com/mehran9675/giode/elements"

	"github.com/xmhrnx/goencode/ui/theme"
)

// selectField is a dropdown whose options can change between frames: the
// codecs on offer change with the container, the encoders with the codec.
type selectField struct {
	sel      *dropdown.Select
	sig      string
	values   []string
	onChange func(string)
	// outside catches clicks anywhere but the open list, which close it.
	outside widget.Clickable
	closing bool
}

// mobile reports whether options open as a bottom sheet. Only phones get
// one: on desktops a narrow settings panel would otherwise count as compact.
var mobile = runtime.GOOS == "android" || runtime.GOOS == "ios"

// Select returns the dropdown identified by key, showing labels for values
// with current selected. onChange receives the chosen value.
func Select(key any, labels, values []string, current string, onChange func(string)) giode.Element {
	f := giode.Keyed("select", key, func() *selectField { return &selectField{} })
	if sig := strings.Join(labels, "\x00"); sig != f.sig || f.sel == nil {
		f.sig = sig
		mode := giode.SheetNever
		if mobile {
			mode = giode.SheetAlways
		}
		f.sel = giode.Input.Select(theme.Field().With(giode.Styles{Width: -1}), labels...).
			Sheet(mode).
			OnChange(func(i int) {
				if f.closing {
					return
				}
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

func (f *selectField) Layout(gtx layout.Context) layout.Dimensions {
	if f.sel.Opened() && !mobile {
		// Registered before the list, so it lies under the list and over
		// everything else, the field included: clicking the field again
		// closes the list too.
		elements.RegisterOverlay(func() elements.Element {
			return elements.Raw(f.layoutOutside)
		})
	}
	return f.sel.Layout(gtx)
}

// layoutOutside is the invisible window-sized layer behind an open list.
func (f *selectField) layoutOutside(gtx layout.Context) layout.Dimensions {
	if f.outside.Clicked(gtx) {
		f.close()
		kit.RequestFrame(gtx)
		return layout.Dimensions{}
	}
	size := gtx.Constraints.Max
	dims := f.outside.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
		return layout.Dimensions{Size: size}
	})
	// Over the custom title bar, a click closes the list rather than
	// dragging the window.
	kit.BlockDrag(gtx, dims.Size)
	return dims
}

// close shuts the list without changing the selection. dropdown.Select has
// no Close, but choosing the selected option closes it; closing keeps that
// from reaching onChange.
func (f *selectField) close() {
	f.closing = true
	f.sel.Choose(f.sel.Selected())
	f.closing = false
}
