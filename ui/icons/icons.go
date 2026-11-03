// Package icons is GoEncode's icon set: simple 24×24 outline glyphs drawn
// for the app, registered by name with giode so any view can use
// giode.Icon("film", ...).
package icons

import (
	_ "embed"

	"github.com/xmhrnx/giode"
)

// Logo is the app icon at 256×256, shown in the header.
//
//go:embed logo.png
var Logo []byte

var paths = map[string]string{
	"plus":          "M12 5v14M5 12h14",
	"x":             "M6 6l12 12M18 6L6 18",
	"film":          "M4 4h16v16H4z M8 4v16 M16 4v16 M4 8h4 M4 12h4 M4 16h4 M16 8h4 M16 12h4 M16 16h4",
	"music":         "M9 18V5l11-2v13 M9 18a3 3 0 1 1-6 0a3 3 0 1 1 6 0 M20 16a3 3 0 1 1-6 0a3 3 0 1 1 6 0",
	"trash":         "M4 7h16 M10 11v6 M14 11v6 M6 7l1 13h10l1-13 M9 7V4h6v3",
	"play":          "M7 4l13 8-13 8z",
	"stop":          "M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18 M9.5 9.5h5v5h-5z",
	"tune":          "M4 6h9 M17 6h3 M15 4v4 M4 12h3 M11 12h9 M9 10v4 M4 18h11 M19 18h1 M17 16v4",
	"cog":           "M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z M12 9a3 3 0 1 0 0 6a3 3 0 1 0 0-6",
	"folder":        "M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z",
	"eye":           "M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12z M12 9a3 3 0 1 0 0 6a3 3 0 1 0 0-6",
	"info":          "M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18 M12 11v5 M12 8h.01",
	"check":         "M5 12l5 5L19 7",
	"alert":         "M12 3L2 20h20z M12 10v4 M12 17h.01",
	"chip":          "M7 7h10v10H7z M10 10h4v4h-4z M9 3v4 M15 3v4 M9 17v4 M15 17v4 M3 9h4 M3 15h4 M17 9h4 M17 15h4",
	"zap":           "M13 2L4 14h7l-1 8 9-12h-7z",
	"layers":        "M12 3l9 5-9 5-9-5z M3 13l9 5 9-5",
	"captions":      "M3 5h18v14H3z M7 13h4 M13 13h4 M7 16h10",
	"bookmark":      "M6 3h12v18l-6-4-6 4z",
	"file":          "M14 3H6v18h12V7z M14 3v4h4",
	"sun":           "M12 8a4 4 0 1 0 0 8a4 4 0 1 0 0-8 M12 2v2 M12 20v2 M2 12h2 M20 12h2 M4.9 4.9l1.4 1.4 M17.7 17.7l1.4 1.4 M4.9 19.1l1.4-1.4 M17.7 6.3l1.4-1.4",
	"moon":          "M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z",
	"sparkle":       "M12 3l1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8z M19 16l.8 2.2L22 19l-2.2.8L19 22l-.8-2.2L16 19l2.2-.8z",
	"device":        "M3 5h18v11H3z M8 20h8 M12 16v4",
	"leaf":          "M5 19c0-8 6-14 15-14 0 9-6 15-14 15 M5 19l7-7",
	"globe":         "M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18 M3 12h18 M12 3c3 3 3 15 0 18 M12 3c-3 3-3 15 0 18",
	"box":           "M3 7l9-4 9 4v10l-9 4-9-4z M3 7l9 4 9-4 M12 11v10",
	"scissors":      "M6 3a3 3 0 1 0 0 6a3 3 0 1 0 0-6 M6 15a3 3 0 1 0 0 6a3 3 0 1 0 0-6 M8.5 7.5L20 19 M8.5 16.5L20 5",
	"image":         "M3 5h18v14H3z M3 16l5-5 4 4 3-3 6 6 M15 9h.01",
	"wave":          "M2 12h2 M6 8v8 M10 5v14 M14 8v8 M18 10v4 M22 12h-2",
	"disc":          "M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18 M12 10a2 2 0 1 0 0 4a2 2 0 1 0 0-4",
	"chevron-down":  "M6 9l6 6 6-6",
	"chevron-right": "M9 6l6 6-6 6",
	"external":      "M14 4h6v6 M20 4l-9 9 M18 14v6H4V6h6",
	"share":         "M12 3v12 M7 8l5-5 5 5 M5 14v6h14v-6",
	"retry":         "M20 12a8 8 0 1 1-2.3-5.7 M20 4v5h-5",
	"clock":         "M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18 M12 7v5l3 3",
	"gauge":         "M4 17a8 8 0 1 1 16 0 M12 17l4-6",
	"upload":        "M12 16V4 M7 9l5-5 5 5 M4 16v4h16v-4",
	"grid":          "M4 4h7v7H4z M13 4h7v7h-7z M4 13h7v7H4z M13 13h7v7h-7z",
	"video":         "M3 6h12v12H3z M15 10l6-3v10l-6-3",
	"speaker":       "M4 9h4l5-4v14l-5-4H4z M17 9a4 4 0 0 1 0 6 M19.5 6.5a7.5 7.5 0 0 1 0 11",
	"download":      "M12 4v12 M7 11l5 5 5-5 M4 20h16",
	"tag":           "M3 12V3h9l9 9-9 9z M8 8h.01",
	"split":         "M4 5h16v14H4z M12 5v14",
	"broom":         "M14 4l6 6 M17 7l-7 7 M4 20l6-6 M10 14l-2 6 M10 14l-6 2",
}

// Register adds the icons to giode's registry. Call it once at startup.
func Register() {
	for name, d := range paths {
		giode.RegisterIcon(name, giode.IconPath{Data: d, StrokeWidth: 1.8})
	}
}
