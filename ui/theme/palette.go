// Package theme holds GoEncode's visual language: palettes for light and
// dark mode, derived surface colors, spacing and type scale.
package theme

import "github.com/mehran9675/giode"

// Brand colors: a violet-to-cyan gradient, the same as the app icon.
const (
	BrandViolet = "#7C5CFF"
	BrandBlue   = "#4F7BFF"
	BrandCyan   = "#22D3EE"
)

var dark = giode.Palette{
	Primary:    "#8B6CFF",
	Secondary:  "#22D3EE",
	Background: "#0A0B10",
	Surface:    "#12141B",
	Text:       "#ECEEF5",
	Border:     "#242833",
	OnPrimary:  "#FFFFFF",
	Muted:      "#8A91A5",
	Accent:     "#22D3EE",
	Success:    "#34D399",
	Warning:    "#FBBF24",
	Danger:     "#F87171",
	Info:       "#60A5FA",
}

var light = giode.Palette{
	Primary:    "#6D4AFF",
	Secondary:  "#0891B2",
	Background: "#F4F5FA",
	Surface:    "#FFFFFF",
	Text:       "#11131C",
	Border:     "#E2E5EE",
	OnPrimary:  "#FFFFFF",
	Muted:      "#5F6679",
	Accent:     "#0891B2",
	Success:    "#059669",
	Warning:    "#D97706",
	Danger:     "#DC2626",
	Info:       "#2563EB",
}

// Install registers GoEncode's palettes for both modes.
func Install() {
	giode.SetPalette(giode.ModeDark, dark)
	giode.SetPalette(giode.ModeLight, light)
}

// Apply switches to a theme preference: "dark", "light" or "system".
func Apply(pref string) {
	switch pref {
	case "dark":
		giode.SetMode(giode.ModeDark)
	case "light":
		giode.SetMode(giode.ModeLight)
	default:
		if d, err := giode.SystemDarkMode(); err == nil && !d {
			giode.SetMode(giode.ModeLight)
		} else {
			giode.SetMode(giode.ModeDark)
		}
	}
}

// IsDark reports whether dark mode is active.
func IsDark() bool { return giode.CurrentMode() == giode.ModeDark }

// Derived surfaces, mixed in sRGB so they follow light/dark switches.
var (
	// Raised is a card on the window background.
	Raised = giode.Mix("theme.text", "theme.surface", 0.025)
	// Sunken is an inset well: inputs, tracks, the file list background.
	Sunken = giode.Mix("theme.text", "theme.background", 0.035)
	// Hover tints a hovered row or card.
	Hover = giode.Mix("theme.text", "theme.surface", 0.06)
	// Subtle is a faint divider.
	Subtle = giode.Mix("theme.text", "theme.surface", 0.08)
	// PrimarySoft is a selection tint.
	PrimarySoft = giode.Mix("theme.primary", "theme.surface", 0.14)
	// PrimaryLine is a selected border.
	PrimaryLine = giode.Mix("theme.primary", "theme.surface", 0.55)
	// TooltipBg is the hover popover.
	TooltipBg = giode.Mix("theme.text", "theme.background", 0.10)
)

// Gradient is the brand gradient at an angle.
func Gradient(angle float32) *giode.LinearGradient {
	return giode.Linear(angle, BrandViolet, BrandBlue, BrandCyan)
}

// Tone returns the soft background and strong foreground of a status tone:
// "success", "warning", "danger", "info", "primary" or "muted".
func Tone(t string) (bg, fg giode.Color) {
	token := giode.Color("theme." + t)
	if t == "" || t == "muted" {
		return Subtle, "theme.muted"
	}
	return giode.Mix(string(token), "theme.surface", 0.16), token
}
