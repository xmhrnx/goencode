// Package views assembles the screens of GoEncode from components, one
// view per file. Views read the store while the frame is built and change
// it only through its actions.
package views

import (
	"context"
	"errors"
	"mime"
	"path/filepath"
	"runtime"

	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/jobs"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/internal/store"
)

// Env is what every view needs: the state, the app (for pickers and
// sharing) and the overlays they can open.
type Env struct {
	Store *store.Store
	App   *giode.App
	Ctx   context.Context

	prefsOpen   bool
	previewOpen bool
}

// Desktop reports whether files have paths and folders can be picked.
func (e *Env) Desktop() bool {
	switch runtime.GOOS {
	case "windows", "darwin", "linux", "freebsd":
		return true
	}
	return false
}

// Framed reports whether GoEncode draws its own window frame: on desktops,
// where the window is frameless and TitleBar replaces the system caption.
func (e *Env) Framed() bool { return e.Desktop() }

// OpenPrefs shows the preferences.
func (e *Env) OpenPrefs() { e.prefsOpen = true }

// OpenPreview shows the preview panel.
func (e *Env) OpenPreview() { e.previewOpen = true }

// AddFiles opens the system file picker.
func (e *Env) AddFiles() {
	go func() {
		if paths, ok, err := platform.NativePickFiles("Add media files", "Video and audio", jobs.MediaExtensions); ok {
			switch {
			case err == nil:
				e.App.Post(func() { e.Store.AddPaths(paths) })
			case !errors.Is(err, platform.ErrPickCanceled):
				e.App.Post(func() { e.Store.Toast(store.ToastError, "Couldn't open the picker: "+err.Error()) })
			}
			return
		}
		// Android's picker takes one file at a time: gioui.org/x/explorer
		// has no multiple choice there.
		files, err := e.App.PickFiles(e.Ctx, giode.PickOptions{Types: pickerTypes(), Multiple: !mobile()})
		if err != nil {
			if !errors.Is(err, giode.ErrCanceled) {
				e.App.Post(func() { e.Store.Toast(store.ToastError, "Couldn't open the picker: "+err.Error()) })
			}
			return
		}
		picked := make([]store.PickedFile, len(files))
		for i, f := range files {
			picked[i] = store.PickedFile{Name: f.Name, Path: f.Path, Body: f.ReadCloser}
		}
		e.Store.AddPicked(picked)
	}()
}

// AddFolder picks a folder and adds every media file inside it.
func (e *Env) AddFolder() {
	go func() {
		dir, err := e.App.PickFolder(e.Ctx, giode.FolderOptions{Title: "Add every media file in a folder"})
		if err != nil {
			return
		}
		e.App.Post(func() { e.Store.AddPaths([]string{dir}) })
	}()
}

// PickOutputDir chooses where results go.
func (e *Env) PickOutputDir() {
	go func() {
		dir, err := e.App.PickFolder(e.Ctx, giode.FolderOptions{Title: "Where should encoded files go?"})
		if err != nil {
			return
		}
		e.App.Post(func() {
			e.Store.Update(func(s *settings.Settings) { s.Output.Dir = dir })
		})
	}()
}

// Reveal shows a result in the file manager, or shares it on phones.
func (e *Env) Reveal(path string) {
	go func() {
		mt := mime.TypeByExtension(filepath.Ext(path))
		if mt == "" {
			mt = "application/octet-stream"
		}
		if err := e.App.Share(e.Ctx, giode.ShareFile{Path: path, MIME: mt}); err != nil {
			e.App.Post(func() { e.Store.Toast(store.ToastError, "Couldn't share: "+err.Error()) })
		}
	}()
}

// Open plays a result in the default app.
func (e *Env) Open(path string) {
	if err := platform.Open(path); err != nil {
		e.Reveal(path)
	}
}

// pickerTypes limits the picker to media on desktops. Phones get every
// file: their picker filters by MIME type, which giode derives from
// extensions with a table Android lacks, so a filter there hides
// everything. ffprobe turns down what isn't media.
func pickerTypes() []string {
	if mobile() {
		return nil
	}
	return jobs.MediaExtensions
}

func mobile() bool { return runtime.GOOS == "android" || runtime.GOOS == "ios" }
