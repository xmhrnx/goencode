package catalog

import (
	"strconv"
	"strings"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// OptionKind is how an option is edited.
type OptionKind int

const (
	OptionChoice OptionKind = iota
	OptionNumber
	OptionToggle
	OptionText
)

// Option is one encoder setting. Its value is passed as -Key:v (or :a).
// An empty value means "leave it to the encoder".
type Option struct {
	Key     string
	Label   string
	Help    string
	Kind    OptionKind
	Choices []Choice
	Min     float64
	Max     float64
	Step    float64
	Unit    string
	// Default is shown as the encoder's own default; it's never passed.
	Default string
	// Simple options show in the basic encoder settings; the rest live
	// under "More options".
	Simple bool
}

// Choice is one value of a choice option.
type Choice struct {
	Value string
	Label string
	Help  string
}

// ChoiceLabel returns the label of a value.
func (o *Option) ChoiceLabel(v string) string {
	for _, c := range o.Choices {
		if c.Value == v {
			return c.Label
		}
	}
	return v
}

// FromAVOption turns an option from ffmpeg's encoder help into an Option,
// for encoders and settings GoEncode has no curated description of. The
// help text is ffmpeg's own description of the option for that encoder.
func FromAVOption(o ffmpeg.AVOption) Option {
	opt := Option{Key: o.Name, Label: prettyKey(o.Name), Help: sentence(o.Description), Default: o.Default}
	if opt.Help == "" {
		opt.Help = "Private option of this encoder."
	}
	switch o.Type {
	case "boolean":
		opt.Kind = OptionToggle
	case "int", "int64", "uint64", "float", "double":
		if len(o.Constants) > 0 {
			opt.Kind = OptionChoice
			for _, c := range o.Constants {
				opt.Choices = append(opt.Choices, Choice{Value: c.Name, Label: c.Name, Help: sentence(c.Description)})
			}
			break
		}
		opt.Kind = OptionNumber
		opt.Min, _ = strconv.ParseFloat(o.Min, 64)
		opt.Max, _ = strconv.ParseFloat(o.Max, 64)
		if o.Type == "float" || o.Type == "double" {
			opt.Step = 0.1
		} else {
			opt.Step = 1
		}
	case "string", "dictionary", "flags", "image_size", "rational", "color", "channel_layout", "sample_fmt", "pixel_fmt", "duration", "binary":
		if len(o.Constants) > 0 && o.Type != "flags" {
			opt.Kind = OptionChoice
			for _, c := range o.Constants {
				opt.Choices = append(opt.Choices, Choice{Value: c.Name, Label: c.Name, Help: sentence(c.Description)})
			}
		} else {
			opt.Kind = OptionText
			if o.Type == "flags" && len(o.Constants) > 0 {
				names := make([]string, len(o.Constants))
				for i, c := range o.Constants {
					names[i] = c.Name
				}
				opt.Help += " Flags: " + strings.Join(names, ", ") + " (combine with +)."
			}
		}
	default:
		opt.Kind = OptionText
	}
	return opt
}

func prettyKey(k string) string {
	k = strings.NewReplacer("_", " ", "-", " ").Replace(k)
	if k == "" {
		return k
	}
	return strings.ToUpper(k[:1]) + k[1:]
}

func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s = strings.ToUpper(s[:1]) + s[1:]
	if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, ")") {
		s += "."
	}
	return s
}
