package components

import (
	"math"
	"strconv"

	"gioui.org/layout"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/slider"

	"github.com/xmhrnx/goencode/ui/theme"
)

// SliderSpec describes a slider's scale.
type SliderSpec struct {
	Min, Max, Step float64
	// Format prints the value in the badge.
	Format func(v float64) string
	// LeftLabel and RightLabel name the ends ("Smaller", "Better").
	LeftLabel, RightLabel string
	// Reverse puts Max on the left, for scales where lower is better.
	Reverse bool
	// Hint is a word or two about the current value ("default").
	Hint string
}

type sliderField struct {
	sl       *slider.Slider
	spec     SliderSpec
	onChange func(float64)
	last     float64
}

// Slider returns the slider identified by key at value on spec's scale.
func Slider(key any, value float64, spec SliderSpec, onChange func(float64)) giode.Element {
	f := giode.Keyed("slider", key, func() *sliderField {
		f := &sliderField{}
		f.sl = giode.Input.Slider(giode.Styles{Color: "theme.primary", Background: theme.Subtle, BorderColor: "theme.primary"}).
			OnChange(func(v float32) {
				val := f.value(float64(v))
				if val != f.last {
					f.last = val
					f.onChange(val)
				}
			})
		return f
	})
	f.spec, f.onChange = spec, onChange
	if !f.sl.Dragging() {
		f.last = value
		f.sl.SetValue(float32(f.fraction(value)))
	}
	return f
}

func (f *sliderField) fraction(v float64) float64 {
	s := f.spec
	if s.Max == s.Min {
		return 0
	}
	fr := (v - s.Min) / (s.Max - s.Min)
	if s.Reverse {
		fr = 1 - fr
	}
	return math.Max(0, math.Min(1, fr))
}

func (f *sliderField) value(fr float64) float64 {
	s := f.spec
	if s.Reverse {
		fr = 1 - fr
	}
	v := s.Min + fr*(s.Max-s.Min)
	if s.Step > 0 {
		v = math.Round(v/s.Step) * s.Step
	}
	return math.Max(s.Min, math.Min(s.Max, v))
}

func (f *sliderField) Layout(gtx layout.Context) layout.Dimensions {
	val := f.last
	text := strconv.FormatFloat(val, 'f', -1, 64)
	if f.spec.Format != nil {
		text = f.spec.Format(val)
	}
	badge := giode.Box(giode.Styles{Background: theme.PrimarySoft, BorderRadius: theme.Pill, Padding: giode.SymmetricInset(9, 3)},
		giode.Text(text, giode.Styles{FontSize: 12, FontWeight: giode.FontWeightSemiBold, Color: "theme.primary"}))
	head := []giode.Element{badge}
	if f.spec.Hint != "" {
		head = append(head, giode.Text(f.spec.Hint, theme.Caption))
	}
	rows := []giode.Element{
		giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter}, head...),
		f.sl,
	}
	if f.spec.LeftLabel != "" || f.spec.RightLabel != "" {
		rows = append(rows, giode.Row(giode.Styles{Justify: giode.JustifySpaceBetween, Width: -1},
			giode.Text(f.spec.LeftLabel, theme.Caption), giode.Text(f.spec.RightLabel, theme.Caption)))
	}
	return giode.Stack(giode.Styles{Gap: 6, Width: -1}, rows...).Layout(gtx)
}
