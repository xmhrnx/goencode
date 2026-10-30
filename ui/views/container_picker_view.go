package views

import (
	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// ContainerPicker chooses the output format. Changing it drops codecs and
// settings the new format can't hold, so every list below only shows what
// fits.
func ContainerPicker(env *Env) giode.Element {
	s := env.Store
	c := catalog.ContainerByID(s.Settings.Container)
	labels, values := []string{}, []string{}
	for _, ct := range catalog.Containers {
		label := ct.Label + "  ·  ." + ct.Ext
		if ct.AudioOnly {
			label += "  ·  audio"
		}
		labels, values = append(labels, label), append(values, ct.ID)
	}
	return giode.Stack(theme.Card().With(giode.Styles{Gap: theme.S2, Width: -1}),
		components.OptionRow(components.InfoLabel("container/"+c.ID, "Format", catalog.HelpContainer(c)),
			components.Select("container", labels, values, c.ID, s.SetContainer)),
		giode.Text(c.Description, theme.Caption),
	)
}
