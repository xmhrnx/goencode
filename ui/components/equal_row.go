package components

import (
	"image"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/mehran9675/giode"
)

type equalRow struct {
	gap      unit.Dp
	children []giode.Element
}

// minHeighter is a child that can grow to a height EqualRow gives it, in
// dp; giode boxes otherwise size to their content.
type minHeighter interface{ SetMinHeight(dp float32) }

// EqualRow lays children side by side in equal widths and gives them all
// the height of the tallest, so a grid of cards lines up whatever their
// text wraps to.
func EqualRow(gap int, children ...giode.Element) giode.Element {
	return &equalRow{gap: unit.Dp(gap), children: children}
}

func (r *equalRow) Layout(gtx layout.Context) layout.Dimensions {
	n := len(r.children)
	if n == 0 {
		return layout.Dimensions{}
	}
	gap := gtx.Dp(r.gap)
	total := gtx.Constraints.Max.X
	w := max(0, (total-gap*(n-1))/n)

	// Measure: lay each child out at its width and throw the drawing away.
	height := 0
	for _, c := range r.children {
		m := op.Record(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Constraints{Min: image.Pt(w, 0), Max: image.Pt(w, gtx.Constraints.Max.Y)}
		dims := c.Layout(cgtx)
		m.Stop()
		height = max(height, dims.Size.Y)
	}

	// Draw every child at that height.
	for i, c := range r.children {
		if m, ok := c.(minHeighter); ok {
			m.SetMinHeight(float32(math.Ceil(float64(height) / float64(gtx.Metric.PxPerDp))))
		}
		off := op.Offset(image.Pt(i*(w+gap), 0)).Push(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Exact(image.Pt(w, height))
		c.Layout(cgtx)
		off.Pop()
	}
	return layout.Dimensions{Size: image.Pt(total, height)}
}
