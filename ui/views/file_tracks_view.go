package views

import (
	"fmt"

	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/ui/components"
	"github.com/xmhrnx/goencode/ui/theme"
)

// FileTracks lists the selected file's audio or subtitle tracks with a
// switch each, to keep exactly the ones wanted. Until one is switched,
// the switches show what the recipe's track rule picks. It's nil when the
// file has fewer than two such tracks.
func FileTracks(env *Env, kind ffmpeg.MediaKind) giode.Element {
	s := env.Store
	it := s.SelectedItem()
	if it == nil || it.Info == nil {
		return nil
	}
	tracks := it.Info.StreamsOf(kind)
	if len(tracks) < 2 {
		return nil
	}
	kept := map[int]bool{}
	if list := it.Picks.Of(kind); list != nil {
		for _, i := range list {
			kept[i] = true
		}
	} else {
		for _, st := range settings.RuleTracks(kind, tracks, s.Settings) {
			kept[st.Index] = true
		}
	}

	noun := "Audio"
	if kind == ffmpeg.KindSubtitle {
		noun = "Subtitle"
	}
	var trailing giode.Element
	sub := it.Name + " · picked by the track rule"
	if it.Picks.Of(kind) != nil {
		sub = it.Name + " · chosen by hand"
		trailing = components.Button(fmt.Sprint("tracks-reset", kind, it.ID), components.ButtonSpec{Label: "Use rule", Icon: "retry", Kind: components.Ghost, Small: true,
			Tooltip: "Keep the tracks the track rule picks", OnClick: func() { s.ResetTracks(it, kind) }})
	}

	rows := make([]giode.Element, 0, len(tracks))
	for n, st := range tracks {
		label := giode.Stack(giode.Styles{Gap: 2},
			giode.Text(fmt.Sprintf("%d. %s", n+1, components.Language(st.Language)), giode.Styles{FontSize: 13, Color: "theme.text"}),
			giode.Text(components.TrackLabel(st), theme.Caption.With(giode.Styles{TextOverflow: giode.TextOverflowEllipsis})))
		index := st.Index
		rows = append(rows, components.OptionRow(label,
			components.Switch(fmt.Sprint("track", kind, it.ID, index), kept[index], func(on bool) { s.PickTrack(it, kind, index, on) })))
	}
	return components.SectionCard("layers", noun+" tracks", sub, trailing, rows...)
}
