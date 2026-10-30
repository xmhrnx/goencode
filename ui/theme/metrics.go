package theme

import "github.com/mehran9675/giode"

// Spacing scale in dp.
const (
	S1 = 4
	S2 = 8
	S3 = 12
	S4 = 16
	S5 = 20
	S6 = 24
	S8 = 32
)

// Corner radii in dp.
const (
	RadiusSm = 8
	RadiusMd = 12
	RadiusLg = 16
	RadiusXl = 20
	Pill     = 999
)

// Layout breakpoints in dp.
const (
	// Compact is the phone layout: one column with tabs.
	Compact = 720
	// Wide puts the settings beside the file list.
	SidePanelWidth = 440
)

// Text styles.
var (
	Display = giode.Styles{FontSize: 22, FontWeight: giode.FontWeightBold, Color: "theme.text"}
	Title   = giode.Styles{FontSize: 16, FontWeight: giode.FontWeightSemiBold, Color: "theme.text"}
	Heading = giode.Styles{FontSize: 13.5, FontWeight: giode.FontWeightSemiBold, Color: "theme.text"}
	Body    = giode.Styles{FontSize: 13.5, Color: "theme.text"}
	Label   = giode.Styles{FontSize: 13, Color: "theme.text"}
	Caption = giode.Styles{FontSize: 12, Color: "theme.muted"}
	Micro   = giode.Styles{FontSize: 10.5, FontWeight: giode.FontWeightSemiBold, Color: "theme.muted", TextTransform: giode.TextTransformUppercase}
	Mono    = giode.Styles{FontSize: 12, FontFamily: "Go Mono", Color: "theme.muted"}
)

// Card is the style of a raised panel.
func Card() giode.Styles {
	return giode.Styles{
		Background:   Raised,
		BorderColor:  "theme.border",
		BorderWidth:  1,
		BorderRadius: RadiusLg,
		Padding:      giode.UniformInset(S4),
	}
}

// Field is the style of input fields and selects.
func Field() giode.Styles {
	return giode.Styles{
		Background:   Sunken,
		BorderColor:  "theme.border",
		BorderRadius: RadiusSm,
		Color:        "theme.text",
		FontSize:     13,
	}
}
