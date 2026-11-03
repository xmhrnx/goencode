package views

import (
	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/components"
)

// ExtrasSettings keeps or strips chapters and metadata.
func ExtrasSettings(env *Env) giode.Element {
	s := env.Store
	st := s.Settings
	return components.SectionCard("bookmark", "Chapters & metadata", "What travels with the streams", nil,
		components.OptionRow(components.InfoLabel("x-chapters", "Keep chapters", catalog.HelpChapters()),
			components.Switch("x-chapters", st.Chapters.Keep, func(on bool) { s.Update(func(x *settings.Settings) { x.Chapters.Keep = on }) })),
		components.OptionRow(components.InfoLabel("x-meta", "Keep metadata", catalog.HelpMetadata()),
			components.Switch("x-meta", st.Metadata.Keep, func(on bool) { s.Update(func(x *settings.Settings) { x.Metadata.Keep = on }) })),
	)
}
