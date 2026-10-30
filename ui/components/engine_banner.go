package components

import (
	"runtime"

	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/theme"
)

// EngineBanner explains a missing or unusable ffmpeg and how to fix it.
func EngineBanner(state store.EngineState, onSettings, onRetry func()) giode.Element {
	switch state {
	case store.EngineLoading:
		return giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter, Padding: giode.SymmetricInset(theme.S4, theme.S2)},
			giode.Spinner(giode.Styles{Width: 14, Color: "theme.primary"}),
			giode.Text("Looking for ffmpeg…", theme.Caption))
	case store.EngineUnsupported:
		return giode.Box(giode.Styles{Padding: giode.SymmetricInset(theme.S4, theme.S2), Width: -1},
			Notice("warning", "alert", "This platform doesn't allow apps to run ffmpeg, so GoEncode can't encode here. Use GoEncode on Android, Windows, macOS or Linux."))
	case store.EngineMissing:
		install := map[string]string{
			"windows": "Install it with  winget install Gyan.FFmpeg  and restart GoEncode, or point GoEncode at ffmpeg.exe in Preferences.",
			"darwin":  "Install it with  brew install ffmpeg  and restart GoEncode, or point GoEncode at it in Preferences.",
			"linux":   "Install it with your package manager (sudo apt install ffmpeg) and restart GoEncode, or point GoEncode at it in Preferences.",
			"android": "This build doesn't include ffmpeg. Install a GoEncode build that bundles it.",
		}[runtime.GOOS]
		if install == "" {
			install = "Install ffmpeg and ffprobe, then set their folder in Preferences."
		}
		return giode.Box(giode.Styles{Padding: giode.SymmetricInset(theme.S4, theme.S2), Width: -1},
			giode.Stack(theme.Card().With(giode.Styles{Gap: theme.S3, Width: -1, BorderColor: "theme.warning"}),
				giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter},
					giode.Icon("alert", giode.Styles{Width: 18, Color: "theme.warning"}),
					giode.Text("ffmpeg wasn't found", theme.Title)),
				giode.Text("GoEncode uses ffmpeg and ffprobe to read and write media. "+install, theme.Body.With(giode.Styles{Color: "theme.muted", LineHeight: 19})),
				giode.Row(giode.Styles{Gap: theme.S2},
					Button("engine-retry", ButtonSpec{Label: "Try again", Icon: "retry", Kind: Secondary, Small: true, OnClick: onRetry}),
					Button("engine-prefs", ButtonSpec{Label: "Set ffmpeg location", Icon: "cog", Kind: Ghost, Small: true, OnClick: onSettings}),
				)))
	}
	return nil
}
