// Command goencode is GoEncode: a cross-platform app that remuxes, encodes
// and transcodes batches of media files with ffmpeg.
package main

import (
	"context"
	_ "embed"
	"image"
	"log"
	"os"
	"runtime"

	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/icons"
	"github.com/xmhrnx/goencode/ui/theme"
	"github.com/xmhrnx/goencode/ui/views"
)

// icon is the window icon; giode build also makes the app icons from it.
//
//go:embed assets/icon.png
var icon []byte

func main() {
	giode.UseGofont()
	theme.Install()
	icons.Register()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var st *store.Store
	app := giode.New(giode.Config{
		Title:     "GoEncode",
		Resizable: true,
		// Desktops get GoEncode's own frame (views.TitleBar); phones have
		// no window frame.
		Frameless:  runtime.GOOS != "android" && runtime.GOOS != "ios",
		Size:       image.Pt(1280, 820),
		MinSize:    image.Pt(380, 560),
		Background: "theme.background",
		Icon:       icon,
		OnFileDrop: func(paths []string) { st.AddPaths(paths) },
		OnSystemThemeChange: func(bool) {
			if st.Prefs.Theme == "" || st.Prefs.Theme == "system" {
				theme.Apply("system")
			}
		},
	})

	st = store.New(ctx, app)
	theme.Apply(st.Prefs.Theme)
	env := &views.Env{Store: st, App: app, Ctx: ctx}
	root := views.NewApp(env)
	// Files passed on the command line ("Open with GoEncode") join the batch.
	if len(os.Args) > 1 {
		st.AddPaths(os.Args[1:])
	}

	if err := app.Run(func() giode.Element { return root }); err != nil {
		log.Fatal(err)
	}
}
