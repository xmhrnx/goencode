package settings

import (
	"fmt"
	"strings"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// Action is what happens to one input stream.
type Action string

const (
	ActCopy    Action = "copy"
	ActEncode  Action = "encode"
	ActConvert Action = "convert" // copy was asked for, but the codec had to change
	ActBurn    Action = "burn"
	ActDrop    Action = "drop"
)

// StreamPlan is the decision for one input stream.
type StreamPlan struct {
	Stream  ffmpeg.Stream
	Action  Action
	Encoder string // for encode and convert
	Codec   string
	Reason  string // why a stream was converted or dropped
}

// Plan is what an encode of one file will do, worked out before running
// it. The UI shows its notes on each file.
type Plan struct {
	Container *catalog.Container
	Video     *StreamPlan // nil: no video in the output
	Audio     []StreamPlan
	Subtitles []StreamPlan
	Burn      *ffmpeg.Stream
	Chapters  bool
	Notes     []string
}

// Warnings lists the conversions and drops the user should know about.
func (p *Plan) Warnings() []string {
	var w []string
	add := func(sp StreamPlan) {
		if sp.Reason != "" && (sp.Action == ActConvert || sp.Action == ActDrop) {
			w = append(w, sp.Reason)
		}
	}
	if p.Video != nil {
		add(*p.Video)
	}
	for _, a := range p.Audio {
		add(a)
	}
	for _, s := range p.Subtitles {
		add(s)
	}
	return append(w, p.Notes...)
}

// Summary describes the plan in a few words: "HEVC → H.264 · 2 audio · 3 subs".
func (p *Plan) Summary() string {
	var parts []string
	if p.Video != nil {
		switch p.Video.Action {
		case ActCopy:
			parts = append(parts, "video copied")
		default:
			parts = append(parts, codecLabel(p.Video.Stream.Codec)+" → "+codecLabel(p.Video.Codec))
		}
	}
	count := func(list []StreamPlan, noun string) {
		n := 0
		for _, s := range list {
			if s.Action != ActDrop {
				n++
			}
		}
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, noun))
		}
	}
	count(p.Audio, "audio")
	count(p.Subtitles, "subs")
	if p.Burn != nil {
		parts = append(parts, "burned subs")
	}
	return strings.Join(parts, " · ")
}

func codecLabel(id string) string {
	l := catalog.Codec(id, "").Label
	if i := strings.Index(l, " / "); i > 0 {
		return l[:i]
	}
	if i := strings.Index(l, " ("); i > 0 {
		return l[:i]
	}
	return l
}

// MakePlan decides what happens to every stream of a file.
func MakePlan(info *ffmpeg.MediaInfo, s Settings, caps *ffmpeg.Capabilities) *Plan {
	c := catalog.ContainerByID(s.Container)
	p := &Plan{Container: c}

	// Video: the first stream that isn't cover art.
	if v, ok := info.Video(); ok && c.Holds(ffmpeg.KindVideo) && s.Video.Mode != ModeDrop {
		sp := StreamPlan{Stream: v}
		switch {
		case s.Video.Mode == ModeEncode:
			sp.Action, sp.Codec, sp.Encoder = ActEncode, s.Video.Codec, s.Video.Encoder
		case c.Accepts(ffmpeg.KindVideo, v.Codec):
			sp.Action, sp.Codec = ActCopy, v.Codec
		case s.Video.CopyFallback && s.Video.Encoder != "":
			sp.Action, sp.Codec, sp.Encoder = ActConvert, s.Video.Codec, s.Video.Encoder
			sp.Reason = fmt.Sprintf("%s video can't go in %s; converted to %s", codecLabel(v.Codec), c.Label, codecLabel(s.Video.Codec))
		default:
			sp.Action = ActDrop
			sp.Reason = fmt.Sprintf("%s video can't go in %s; dropped", codecLabel(v.Codec), c.Label)
		}
		if sp.Action != ActDrop {
			p.Video = &sp
		} else {
			p.Notes = append(p.Notes, sp.Reason)
		}
	}

	// Audio.
	if c.Holds(ffmpeg.KindAudio) && s.Audio.Mode != ModeDrop {
		for _, a := range selectTracks(info.StreamsOf(ffmpeg.KindAudio), s.Audio.Tracks, s.Audio.Languages) {
			sp := StreamPlan{Stream: a}
			needsEncode := s.Audio.Mode == ModeEncode || s.Audio.Loudnorm
			switch {
			case needsEncode:
				sp.Action, sp.Codec, sp.Encoder = ActEncode, s.Audio.Codec, s.Audio.Encoder
			case c.Accepts(ffmpeg.KindAudio, a.Codec):
				sp.Action, sp.Codec = ActCopy, a.Codec
			case s.Audio.CopyFallback && s.Audio.Encoder != "":
				sp.Action, sp.Codec, sp.Encoder = ActConvert, s.Audio.Codec, s.Audio.Encoder
				sp.Reason = fmt.Sprintf("%s audio can't go in %s; converted to %s", codecLabel(a.Codec), c.Label, codecLabel(s.Audio.Codec))
			default:
				sp.Action = ActDrop
				sp.Reason = fmt.Sprintf("%s audio can't go in %s; dropped", codecLabel(a.Codec), c.Label)
			}
			p.Audio = append(p.Audio, sp)
		}
	}

	// Subtitles.
	subs := info.StreamsOf(ffmpeg.KindSubtitle)
	switch s.Subtitles.Mode {
	case SubsBurn:
		if b, ok := burnTrack(subs, s.Subtitles.BurnLanguage); ok && p.Video != nil && p.Video.Action != ActCopy {
			p.Burn = &b
		} else if len(subs) > 0 && p.Video != nil && p.Video.Action == ActCopy {
			p.Notes = append(p.Notes, "Subtitles can't be burned in while the video is copied")
		}
	case SubsKeep:
		textTargets := TextSubtitleCodecs(caps, c)
		selected := selectTracks(subs, s.Subtitles.Tracks, s.Subtitles.Languages)
		if len(selected) > 0 && !c.Holds(ffmpeg.KindSubtitle) {
			p.Notes = append(p.Notes, fmt.Sprintf("%s can't store subtitles; %d track(s) left out", c.Label, len(selected)))
			break
		}
		for _, st := range selected {
			sp := StreamPlan{Stream: st}
			switch {
			case s.Subtitles.Codec != "" && st.TextSubtitle():
				sp.Action, sp.Codec = ActEncode, s.Subtitles.Codec
				if st.Codec == sp.Codec {
					sp.Action = ActCopy
				}
			case c.Accepts(ffmpeg.KindSubtitle, st.Codec):
				sp.Action, sp.Codec = ActCopy, st.Codec
			case st.TextSubtitle() && len(textTargets) > 0:
				target := PreferredCodec(caps, c, ffmpeg.KindSubtitle)
				if !contains(textTargets, target) {
					target = textTargets[0]
				}
				sp.Action, sp.Codec = ActConvert, target
				sp.Reason = fmt.Sprintf("%s subtitles converted to %s for %s", codecLabel(st.Codec), codecLabel(target), c.Label)
			default:
				sp.Action = ActDrop
				sp.Reason = fmt.Sprintf("%s subtitles are pictures %s can't store; dropped", codecLabel(st.Codec), c.Label)
			}
			p.Subtitles = append(p.Subtitles, sp)
		}
	}

	p.Chapters = s.Chapters.Keep && c.Chapters && info.Chapters > 0
	if s.Chapters.Keep && !c.Chapters && info.Chapters > 0 {
		p.Notes = append(p.Notes, c.Label+" can't store chapters")
	}
	return p
}

// selectTracks applies a track rule.
func selectTracks(streams []ffmpeg.Stream, rule TrackRule, languages string) []ffmpeg.Stream {
	switch rule {
	case TracksFirst:
		if len(streams) > 1 {
			return streams[:1]
		}
	case TracksLanguages:
		want := parseLanguages(languages)
		if len(want) == 0 {
			return streams
		}
		var out []ffmpeg.Stream
		for _, s := range streams {
			lang := strings.ToLower(s.Language)
			if lang == "" {
				lang = "und"
			}
			if want[lang] {
				out = append(out, s)
			}
		}
		return out
	}
	return streams
}

func parseLanguages(s string) map[string]bool {
	m := map[string]bool{}
	for _, f := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		m[f] = true
	}
	return m
}

// burnTrack picks the subtitle to burn: matching the language if given,
// preferring forced, then default, then the first.
func burnTrack(subs []ffmpeg.Stream, language string) (ffmpeg.Stream, bool) {
	if len(subs) == 0 {
		return ffmpeg.Stream{}, false
	}
	pool := subs
	if language != "" {
		pool = selectTracks(subs, TracksLanguages, language)
		if len(pool) == 0 {
			return ffmpeg.Stream{}, false
		}
	}
	for _, s := range pool {
		if s.Forced {
			return s, true
		}
	}
	for _, s := range pool {
		if s.Default {
			return s, true
		}
	}
	return pool[0], true
}
