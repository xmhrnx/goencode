package components

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
)

// OptionControl renders one encoder option as a labeled row: a dropdown
// for choices, a number field, an on/off choice or a text field, with the
// option's codec-specific explanation on hover. scope keeps the widgets of
// different encoders apart.
func OptionControl(scope string, o catalog.Option, value string, onChange func(string)) giode.Element {
	key := scope + "/" + o.Key
	help := o.Help
	if o.Default != "" && o.Kind != catalog.OptionToggle {
		help += " Default: " + o.ChoiceLabel(o.Default) + "."
	}
	label := InfoLabel(key, o.Label, help)

	var control giode.Element
	switch o.Kind {
	case catalog.OptionChoice:
		auto := "Auto"
		if o.Default != "" {
			auto = "Default (" + o.ChoiceLabel(o.Default) + ")"
		}
		labels := []string{auto}
		values := []string{""}
		for _, c := range o.Choices {
			if c.Value == o.Default {
				continue
			}
			labels = append(labels, c.Label)
			values = append(values, c.Value)
		}
		control = Select(key, labels, values, value, onChange)
	case catalog.OptionToggle:
		control = Select(key, []string{"Default", "On", "Off"}, []string{"", "1", "0"}, value, onChange)
	case catalog.OptionNumber:
		placeholder := "Default"
		if o.Default != "" {
			placeholder = o.Default
		}
		lo, hi := o.Min, o.Max
		if hi > 1e9 || lo < -1e9 {
			lo, hi = 0, 0 // ffmpeg's INT_MAX ranges: don't bother validating
		}
		if hi > lo && placeholder == "Default" {
			placeholder = trimFloat(lo) + " – " + trimFloat(hi)
		}
		control = Number(key, value, placeholder, o.Unit, lo, hi, onChange)
	default:
		placeholder := "Default"
		if o.Default != "" {
			placeholder = o.Default
		}
		control = TextInput(key, value, placeholder, onChange)
	}
	return OptionRow(label, control)
}

func trimFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
