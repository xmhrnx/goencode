package views

import (
	"path/filepath"

	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// OutputSettings chooses where results go and how they're named.
func OutputSettings(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	c := catalog.ContainerByID(st.Container)

	var rows []giode.Element
	if env.Desktop() {
		where := "Next to each source file"
		if st.Output.Dir != "" {
			where = st.Output.Dir
		}
		buttons := []giode.Element{components.Button("out-pick", components.ButtonSpec{Label: "Choose", Icon: "folder", Kind: components.Secondary, Small: true, OnClick: env.PickOutputDir})}
		if st.Output.Dir != "" {
			buttons = append(buttons, components.Button("out-reset", components.ButtonSpec{Icon: "x", Kind: components.Ghost, Small: true,
				Tooltip: "Save next to the source files", OnClick: func() { s.Update(func(x *settings.Settings) { x.Output.Dir = "" }) }}))
		}
		rows = append(rows, giode.Stack(giode.Styles{Gap: theme.S2, Width: -1},
			components.InfoLabel("out-dir", "Folder", catalog.HelpOutputFolder()),
			giode.Row(giode.Styles{Gap: theme.S2, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap},
				giode.Box(giode.Styles{FlexGrow: 1, Padding: giode.SymmetricInset(theme.S3, theme.S2), Background: theme.Sunken, BorderRadius: theme.RadiusSm},
					giode.Text(where, giode.Styles{FontSize: 12, Color: "theme.text", MaxLines: 1})),
				giode.Row(giode.Styles{Gap: 4, Wrap: giode.WrapNoWrap}, buttons...))))
	} else {
		msg := "Finished files are kept in the app; share or save them from each file's card."
		if dir := platform.PublicDir(); dir != "" {
			msg = "Finished files go to Download/GoEncode, where the Files app shows them. Play or share them from each file's card."
		}
		rows = append(rows, components.Notice("info", "share", msg))
	}

	example := "movie" + st.Output.Suffix + "." + c.Ext
	rows = append(rows,
		components.OptionRow(components.InfoLabel("out-suffix", "Name suffix", catalog.HelpSuffix()),
			components.TextInput("out-suffix", st.Output.Suffix, "none", func(v string) { s.Update(func(x *settings.Settings) { x.Output.Suffix = v }) })),
		giode.Text("Example: "+filepath.Base(example), theme.Caption),
		components.OptionRow(components.InfoLabel("out-overwrite", "Overwrite existing", catalog.HelpOverwrite()),
			components.Switch("out-overwrite", st.Output.Overwrite, func(on bool) { s.Update(func(x *settings.Settings) { x.Output.Overwrite = on }) })),
	)
	return components.SectionCard("download", "Output", "Where results are saved", nil, rows...)
}
