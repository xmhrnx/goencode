package ffmpeg

import (
	"context"
	"strings"
	"sync"
)

// MediaKind is what a codec carries.
type MediaKind string

const (
	KindVideo    MediaKind = "video"
	KindAudio    MediaKind = "audio"
	KindSubtitle MediaKind = "subtitle"
)

// Codec is one entry of `ffmpeg -codecs` that this ffmpeg can encode.
type Codec struct {
	ID          string // "h264"
	Description string // "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10"
	Kind        MediaKind
	Encoders    []string // "libx264", "h264_nvenc", ...
	Lossless    bool
	Lossy       bool
}

// Capabilities describe what the located ffmpeg can do.
type Capabilities struct {
	Binaries Binaries
	Codecs   map[string]*Codec // by codec ID; only codecs with an encoder
	// EncoderCodec maps an encoder name to its codec ID.
	EncoderCodec map[string]string
	HWAccels     []string // `ffmpeg -hwaccels`, for decoding

	mu       sync.Mutex
	working  map[string]bool // hardware encoders that passed a test encode
	tested   bool
	encHelp  map[string]*EncoderHelp
	helpWait map[string]chan struct{}
}

// Detect asks ffmpeg for its codecs and hardware decoders.
func Detect(ctx context.Context, b Binaries) (*Capabilities, error) {
	out, err := command(ctx, b.FFmpeg, "-hide_banner", "-codecs").Output()
	if err != nil {
		return nil, err
	}
	c := NewCapabilities(b, ParseCodecs(string(out)))
	if hw, err := command(ctx, b.FFmpeg, "-hide_banner", "-hwaccels").Output(); err == nil {
		for _, line := range strings.Split(string(hw), "\n")[1:] {
			if s := strings.TrimSpace(line); s != "" {
				c.HWAccels = append(c.HWAccels, s)
			}
		}
	}
	return c, nil
}

// NewCapabilities builds capabilities from a codec table. Hardware encoders
// listed in workingHW count as tested and working.
func NewCapabilities(b Binaries, codecs map[string]*Codec, workingHW ...string) *Capabilities {
	c := &Capabilities{
		Binaries:     b,
		Codecs:       codecs,
		EncoderCodec: map[string]string{},
		working:      map[string]bool{},
		encHelp:      map[string]*EncoderHelp{},
		helpWait:     map[string]chan struct{}{},
	}
	for id, cd := range codecs {
		for _, e := range cd.Encoders {
			c.EncoderCodec[e] = id
		}
	}
	for _, e := range workingHW {
		c.working[e] = true
	}
	return c
}

// ParseCodecs reads the table of `ffmpeg -codecs`:
//
//	DEV.LS h264    H.264 / AVC (decoders: h264 h264_qsv) (encoders: libx264 h264_nvenc)
func ParseCodecs(out string) map[string]*Codec {
	codecs := map[string]*Codec{}
	started := false
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(strings.TrimSpace(line), "---") {
			started = true
			continue
		}
		if !started || len(line) < 9 {
			continue
		}
		flags := strings.TrimSpace(line)[:6]
		rest := strings.TrimSpace(strings.TrimSpace(line)[6:])
		if flags[1] != 'E' {
			continue
		}
		id, desc, _ := strings.Cut(rest, " ")
		desc = strings.TrimSpace(desc)
		cd := &Codec{ID: id, Lossy: flags[4] == 'L', Lossless: flags[5] == 'S'}
		switch flags[2] {
		case 'V':
			cd.Kind = KindVideo
		case 'A':
			cd.Kind = KindAudio
		case 'S':
			cd.Kind = KindSubtitle
		default:
			continue
		}
		if i := strings.Index(desc, "(encoders:"); i >= 0 {
			list := desc[i+len("(encoders:"):]
			list, _, _ = strings.Cut(list, ")")
			cd.Encoders = strings.Fields(list)
			desc = desc[:i]
		} else {
			cd.Encoders = []string{id}
		}
		if i := strings.Index(desc, "(decoders:"); i >= 0 {
			desc = desc[:i]
		}
		cd.Description = strings.TrimSpace(desc)
		codecs[id] = cd
	}
	return codecs
}

// HasEncoder reports whether ffmpeg was built with the encoder.
func (c *Capabilities) HasEncoder(name string) bool {
	_, ok := c.EncoderCodec[name]
	return ok
}

// CodecsOfKind lists the encodable codecs of one kind.
func (c *Capabilities) CodecsOfKind(k MediaKind) []*Codec {
	var list []*Codec
	for _, cd := range c.Codecs {
		if cd.Kind == k {
			list = append(list, cd)
		}
	}
	return list
}
