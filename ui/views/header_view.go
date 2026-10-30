package views

import (
	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// Header is the top bar on phones and tablets, which have no window frame
// to draw: the brand, theme and preferences, and the button that adds
// files. Desktops show the same inside TitleBar.
func Header(env *Env, compact bool) giode.Element {
	if compact {
		return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: theme.S2, Wrap: giode.WrapNoWrap,
			Padding: giode.Inset{Top: theme.S3, Right: theme.S3, Bottom: theme.S2, Left: theme.S4}},
			giode.Box(giode.Styles{FlexGrow: 1}, components.Logo(30, true)),
			giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, headerActions(env, true)...),
		)
	}
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: theme.S4, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(theme.S5, theme.S3)},
		giode.Box(giode.Styles{FlexGrow: 1}, components.Logo(34, true)),
		giode.Box(giode.Styles{Width: 250}, ModeSwitch(env, "header")),
		giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, headerActions(env, false)...),
	)
}

// headerActions are the theme and preferences buttons and the ones that
// add files, shared by Header and TitleBar.
func headerActions(env *Env, compact bool) []giode.Element {
	s := env.Store
	small := env.Framed()
	themeIcon, themeTip := "moon", "Switch to dark mode"
	if theme.IsDark() {
		themeIcon, themeTip = "sun", "Switch to light mode"
	}
	actions := []giode.Element{
		components.Button("theme", components.ButtonSpec{Icon: themeIcon, Kind: components.Ghost, Tooltip: themeTip, OnClick: func() {
			next := "dark"
			if theme.IsDark() {
				next = "light"
			}
			p := s.Prefs
			p.Theme = next
			s.SetPrefs(p)
			theme.Apply(next)
		}}),
		components.Button("prefs", components.ButtonSpec{Icon: "cog", Kind: components.Ghost, Tooltip: "Preferences", OnClick: env.OpenPrefs}),
	}
	if compact {
		return append(actions, components.Button("add-compact", components.ButtonSpec{Icon: "plus", Kind: components.Primary,
			Tooltip: "Add files", OnClick: env.AddFiles}))
	}
	if env.Desktop() {
		actions = append(actions, components.Button("add-folder", components.ButtonSpec{Label: "Folder", Icon: "folder", Kind: components.Secondary, Small: small,
			Tooltip: "Add every media file in a folder", OnClick: env.AddFolder}))
	}
	return append(actions, components.Button("add", components.ButtonSpec{Label: "Add files", Icon: "plus", Kind: components.Primary, Small: small,
		OnClick: env.AddFiles}))
}
