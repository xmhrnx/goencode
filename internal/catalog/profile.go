package catalog

import (
	"strings"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// RateControl is one way an encoder decides how many bits to spend.
type RateControl struct {
	ID    string // "crf"
	Label string // "Constant quality"
	Help  string
	// Bitrate controls take a value in kb/s passed as -b:v / -b:a; the
	// others pass the value with Flag.
	Bitrate bool
	Flag    string // "crf", "cq", "q", "global_quality", ...
	// AlsoFlags get the same value as Flag (AMF's per-frame-type QPs).
	AlsoFlags []string
	Min       float64
	Max       float64
	Step      float64
	Default   float64
	// LowerIsBetter: a lower value means higher quality (CRF, QP).
	LowerIsBetter bool
	// Args are passed along with the value, such as "-rc vbr -b:v 0" for
	// NVENC's constant quality.
	Args []string
	// Marks annotate notable values on the quality scale.
	Marks []Mark
	// None means the encoder takes no rate setting (lossless codecs).
	None bool
}

// Mark labels a value on a rate control's scale.
type Mark struct {
	Value float64
	Label string
}

// Profile is everything GoEncode knows about one encoder.
type Profile struct {
	Encoder  string
	Label    string // "x264"
	Kind     ffmpeg.MediaKind
	Summary  string // shown when hovering the encoder's name
	Hardware string // family name for hardware encoders

	RateControls []RateControl
	Options      []Option
	// Managed lists AVOption names GoEncode sets itself, left out of the
	// "all encoder options" list.
	Managed []string
	// Curated reports whether the profile was written by hand, as opposed
	// to built from ffmpeg's help.
	Curated bool
}

// Option returns a curated option by key.
func (p *Profile) Option(key string) *Option {
	for i := range p.Options {
		if p.Options[i].Key == key {
			return &p.Options[i]
		}
	}
	return nil
}

// RateControl returns a rate control by ID, or the first one.
func (p *Profile) RateControl(id string) *RateControl {
	for i := range p.RateControls {
		if p.RateControls[i].ID == id {
			return &p.RateControls[i]
		}
	}
	if len(p.RateControls) == 0 {
		return nil
	}
	return &p.RateControls[0]
}

// IsManaged reports whether an AVOption is already covered by the curated
// settings.
func (p *Profile) IsManaged(name string) bool {
	if p.Option(name) != nil {
		return true
	}
	for _, m := range p.Managed {
		if m == name {
			return true
		}
	}
	return false
}

var profiles = map[string]*Profile{}

func register(ps ...Profile) {
	for i := range ps {
		p := ps[i]
		p.Curated = true
		if fam, ok := ffmpeg.HardwareOf(p.Encoder); ok && p.Hardware == "" {
			p.Hardware = fam.Name
		}
		profiles[p.Encoder] = &p
	}
}

// ProfileFor returns the profile of an encoder, generating a generic one for
// encoders without a curated profile.
func ProfileFor(encoder string, kind ffmpeg.MediaKind, codecDesc string) *Profile {
	if p, ok := profiles[encoder]; ok {
		return p
	}
	if p := hardwareProfile(encoder); p != nil {
		profiles[encoder] = p
		return p
	}
	p := &Profile{Encoder: encoder, Label: encoder, Kind: kind}
	if fam, ok := ffmpeg.HardwareOf(encoder); ok {
		p.Hardware = fam.Name
	}
	name := codecDesc
	if name == "" {
		name = encoder
	}
	p.Summary = "ffmpeg's " + encoder + " encoder for " + name + ". GoEncode lists its own options below, with ffmpeg's descriptions."
	switch kind {
	case ffmpeg.KindVideo:
		p.RateControls = []RateControl{
			{ID: "bitrate", Label: "Target bitrate", Bitrate: true, Min: 100, Max: 100000, Step: 100, Default: 5000,
				Help: "Average number of kilobits per second " + encoder + " aims for. Higher values mean better quality and bigger files."},
			{ID: "qscale", Label: "Quantizer scale", Flag: "q:v", Min: 1, Max: 31, Step: 1, Default: 4, LowerIsBetter: true,
				Help: "Fixed quantizer (-q:v) for encoders that support it: 1 is the best quality, 31 the worst. Ignored by encoders without a quantizer scale."},
			{ID: "default", Label: "Encoder default", None: true,
				Help: "Pass no rate setting and let " + encoder + " use its built-in default."},
		}
	case ffmpeg.KindAudio:
		p.RateControls = []RateControl{
			{ID: "bitrate", Label: "Bitrate", Bitrate: true, Min: 8, Max: 1536, Step: 8, Default: 192,
				Help: "Kilobits per second of audio. Lossless and PCM encoders ignore it."},
			{ID: "default", Label: "Encoder default", None: true,
				Help: "Pass no bitrate and let " + encoder + " decide."},
		}
	}
	p.Managed = []string{"b", "q", "qscale"}
	profiles[encoder] = p
	return p
}

// EncoderLabel returns a short human name for an encoder.
func EncoderLabel(encoder string) string {
	if p, ok := profiles[encoder]; ok {
		return p.Label
	}
	if fam, ok := ffmpeg.HardwareOf(encoder); ok {
		return fam.Name
	}
	return strings.TrimPrefix(encoder, "lib")
}
