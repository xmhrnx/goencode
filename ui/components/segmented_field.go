package components

import (
	"strings"

	"gioui.org/layout"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/segmented"

	"github.com/xmhrnx/goencode/ui/theme"
)

type segmentedField struct {
	seg      *segmented.Control
	sig      string
	values   []string
	onChange func(string)
}

// Segmented returns the segmented control identified by key, choosing
// among values shown as labels, optionally with icon names.
func Segmented(key any, labels, values, icons []string, current string, onChange func(string)) giode.Element {
	f := giode.Keyed("segmented", key, func() *segmentedField { return &segmentedField{} })
	if sig := strings.Join(labels, "\x00"); sig != f.sig || f.seg == nil {
		f.sig = sig
		f.seg = giode.Input.Segmented(labels...).OnChange(func(i int) {
			if i >= 0 && i < len(f.values) {
				f.onChange(f.values[i])
			}
		})
		f.seg.Styles(giode.Styles{Width: -1, FontSize: 12.5, Background: theme.Sunken, BorderColor: theme.Raised, Color: "theme.text"})
		if len(icons) == len(labels) {
			els := make([]giode.Element, len(icons))
			for i, name := range icons {
				els[i] = giode.Icon(name, giode.Styles{Width: 15})
			}
			f.seg.Icons(els...)
		}
	}
	f.values, f.onChange = values, onChange
	for i, v := range values {
		if v == current && f.seg.Selected() != i {
			f.seg.SetSelected(i)
		}
	}
	return f
}

func (f *segmentedField) Layout(gtx layout.Context) layout.Dimensions { return f.seg.Layout(gtx) }
