package views

import (
	"fmt"

	"gioui.org/layout"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/sheet"

	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// App is the root of the window. On wide screens the files sit on the left
// and the settings on the right; on phones a switch flips between them.
type App struct {
	env      *Env
	title    *TitleBar
	files    *FileList
	settings *SettingsPanel
	preview  *PreviewPanel
	prefs    *PrefsPanel
	toasts   *giode.ToasterElement
	prefsSh  *sheet.Sheet
	pvSheet  *sheet.Sheet
	tab      string
}

// NewApp builds the root view and routes the store's messages to toasts.
func NewApp(env *Env) *App {
	a := &App{env: env, tab: "files"}
	if env.Framed() {
		a.title = NewTitleBar(env)
	}
	a.files = NewFileList(env)
	a.settings = NewSettingsPanel(env)
	a.preview = NewPreviewPanel(env)
	a.prefs = NewPrefsPanel(env)
	a.toasts = giode.Toaster().Placement(giode.ToastBottomCenter)
	a.prefsSh = giode.Sheet(&env.prefsOpen, gutter(a.prefs.View), giode.Styles{Background: "theme.background", Padding: giode.SymmetricInset(theme.S5, theme.S3)}).
		Title("Preferences").Adaptive(true)
	a.pvSheet = giode.Sheet(&env.previewOpen, gutter(a.preview.View), giode.Styles{Background: "theme.background", Padding: giode.SymmetricInset(theme.S5, theme.S3)}).
		Title("Preview").Adaptive(true).Dismissible(true)
	kinds := map[store.ToastKind]giode.ToastKind{
		store.ToastInfo: giode.ToastInfo, store.ToastSuccess: giode.ToastSuccess,
		store.ToastWarning: giode.ToastWarning, store.ToastError: giode.ToastError,
	}
	env.Store.Toast = func(k store.ToastKind, msg string) { a.toasts.Show(kinds[k], msg) }
	return a
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	frame := giode.Styles{Width: -1, Height: -1, Background: "theme.background"}
	if a.title != nil && !a.env.App.Maximized() {
		// A frameless window draws its own edge.
		frame.BorderColor, frame.BorderWidth = "theme.border", 1
	}
	return giode.Box(frame,
		giode.Responsive(
			giode.Breakpoint{MaxWidth: theme.Compact, Name: "compact", View: a.compact},
			giode.Breakpoint{View: a.wide},
		),
		a.prefsSh, a.pvSheet, a.toasts,
	).Layout(gtx)
}

func (a *App) wide() giode.Element {
	env := a.env
	return giode.Stack(giode.Styles{Width: -1, Height: -1},
		a.top(false),
		giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1, Height: -1}, a.files.Compact(false)),
			giode.Stack(giode.Styles{Width: theme.SidePanelWidth, Height: -1, Background: theme.Sunken, BorderColor: "theme.border", BorderWidth: 1},
				giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, a.settings.Compact(false)),
				ActionBar(env, false)),
		),
	)
}

func (a *App) compact() giode.Element {
	env := a.env
	filesLabel := "Files"
	if n := len(env.Store.Items); n > 0 {
		filesLabel = fmt.Sprintf("Files · %d", n)
	}
	var body giode.Element = a.files.Compact(true)
	if a.tab == "settings" {
		body = a.settings.Compact(true)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1},
		a.top(true),
		giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Top: theme.S2, Right: theme.S3, Bottom: theme.S1, Left: theme.S3}},
			components.Segmented("tab", []string{filesLabel, "Settings"}, []string{"files", "settings"}, []string{"film", "tune"}, a.tab,
				func(v string) { a.tab = v })),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, body),
		ActionBar(env, true),
	)
}

// top is the custom title bar on desktops and the header on phones.
func (a *App) top(compact bool) giode.Element {
	if a.title != nil {
		return a.title.Compact(compact)
	}
	if compact {
		return Header(a.env, true)
	}
	return giode.Stack(giode.Styles{Width: -1}, Header(a.env, false), giode.Divider(giode.Styles{Color: "theme.border"}))
}

// gutter keeps a sheet's content clear of its scroll bar.
func gutter(view func() giode.Element) func() giode.Element {
	return func() giode.Element {
		return giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Right: theme.S3, Bottom: theme.S4}}, view())
	}
}
