// Package settings is the encoding recipe applied to every file in the
// batch: what to do with video, audio, subtitles, chapters and metadata,
// how to keep it valid for the chosen container, and the ffmpeg arguments
// it turns into.
package settings

// StreamMode is what happens to a kind of stream.
type StreamMode string

const (
	ModeCopy   StreamMode = "copy"
	ModeEncode StreamMode = "encode"
	ModeDrop   StreamMode = "drop"
)

// TrackRule selects which tracks of a kind are kept.
type TrackRule string

const (
	TracksAll       TrackRule = "all"
	TracksFirst     TrackRule = "first"
	TracksLanguages TrackRule = "languages"
)

// SubtitleMode is what happens to subtitles.
type SubtitleMode string

const (
	SubsKeep SubtitleMode = "keep"
	SubsBurn SubtitleMode = "burn"
	SubsDrop SubtitleMode = "drop"
)

// Settings is the full recipe.
type Settings struct {
	Container string    `json:"container"`
	Video     Video     `json:"video"`
	Audio     Audio     `json:"audio"`
	Subtitles Subtitles `json:"subtitles"`
	Chapters  Chapters  `json:"chapters"`
	Metadata  Metadata  `json:"metadata"`
	Output    Output    `json:"output"`
}

// Video settings.
type Video struct {
	Mode        StreamMode `json:"mode"`
	Codec       string     `json:"codec"`
	Encoder     string     `json:"encoder"`
	RateControl string     `json:"rateControl"`
	Quality     float64    `json:"quality"`
	Bitrate     int        `json:"bitrate"` // kb/s
	MaxBitrate  int        `json:"maxBitrate"`
	Height      int        `json:"height"`    // 0 keeps the source size
	FrameRate   float64    `json:"frameRate"` // 0 keeps the source rate
	PixelFormat string     `json:"pixelFormat"`
	HWDecode    bool       `json:"hwDecode"`
	// CopyFallback converts streams that can't be copied into the
	// container instead of dropping them.
	CopyFallback bool `json:"copyFallback"`
	// Options are encoder options by encoder, so switching encoders and
	// back keeps what was set.
	Options map[string]map[string]string `json:"options"`
}

// Audio settings.
type Audio struct {
	Mode         StreamMode                   `json:"mode"`
	Tracks       TrackRule                    `json:"tracks"`
	Languages    string                       `json:"languages"`
	Codec        string                       `json:"codec"`
	Encoder      string                       `json:"encoder"`
	RateControl  string                       `json:"rateControl"`
	Quality      float64                      `json:"quality"`
	Bitrate      int                          `json:"bitrate"`
	Channels     int                          `json:"channels"`   // 0 keeps the source
	SampleRate   int                          `json:"sampleRate"` // 0 keeps the source
	Loudnorm     bool                         `json:"loudnorm"`
	CopyFallback bool                         `json:"copyFallback"`
	Options      map[string]map[string]string `json:"options"`
}

// Subtitles settings.
type Subtitles struct {
	Mode      SubtitleMode `json:"mode"`
	Tracks    TrackRule    `json:"tracks"`
	Languages string       `json:"languages"`
	// Codec converts every text subtitle to this codec; empty copies them
	// when the container allows and converts only when it must.
	Codec string `json:"codec"`
	// BurnLanguage picks the track to burn in; empty takes the forced, then
	// default, then first track.
	BurnLanguage string `json:"burnLanguage"`
}

// Chapters settings.
type Chapters struct {
	Keep bool `json:"keep"`
}

// Metadata settings.
type Metadata struct {
	Keep bool `json:"keep"`
}

// Output settings.
type Output struct {
	Dir       string `json:"dir"` // empty: next to the source
	Suffix    string `json:"suffix"`
	Overwrite bool   `json:"overwrite"`
}

// Clone returns a deep copy.
func (s Settings) Clone() Settings {
	s.Video.Options = cloneOptions(s.Video.Options)
	s.Audio.Options = cloneOptions(s.Audio.Options)
	return s
}

func cloneOptions(m map[string]map[string]string) map[string]map[string]string {
	out := make(map[string]map[string]string, len(m))
	for enc, opts := range m {
		o := make(map[string]string, len(opts))
		for k, v := range opts {
			o[k] = v
		}
		out[enc] = o
	}
	return out
}

// VideoOption returns an option of the current video encoder.
func (s *Settings) VideoOption(key string) string { return s.Video.Options[s.Video.Encoder][key] }

// SetVideoOption sets (or with "" clears) an option of the video encoder.
func (s *Settings) SetVideoOption(key, val string) {
	s.Video.Options = setOption(s.Video.Options, s.Video.Encoder, key, val)
}

// AudioOption returns an option of the current audio encoder.
func (s *Settings) AudioOption(key string) string { return s.Audio.Options[s.Audio.Encoder][key] }

// SetAudioOption sets (or with "" clears) an option of the audio encoder.
func (s *Settings) SetAudioOption(key, val string) {
	s.Audio.Options = setOption(s.Audio.Options, s.Audio.Encoder, key, val)
}

func setOption(m map[string]map[string]string, enc, key, val string) map[string]map[string]string {
	if m == nil {
		m = map[string]map[string]string{}
	}
	if m[enc] == nil {
		m[enc] = map[string]string{}
	}
	if val == "" {
		delete(m[enc], key)
	} else {
		m[enc][key] = val
	}
	return m
}

// Default returns GoEncode's starting recipe: H.264 in MP4, which any
// device plays.
func Default() Settings {
	return Settings{
		Container: "mp4",
		Video: Video{Mode: ModeEncode, Codec: "h264", Encoder: "libx264", RateControl: "crf", Quality: 21, Bitrate: 6000,
			CopyFallback: true, Options: map[string]map[string]string{"libx264": {"preset": "medium"}}},
		Audio: Audio{Mode: ModeEncode, Tracks: TracksAll, Codec: "aac", Encoder: "aac", RateControl: "bitrate", Bitrate: 160,
			CopyFallback: true, Options: map[string]map[string]string{}},
		Subtitles: Subtitles{Mode: SubsKeep, Tracks: TracksAll},
		Chapters:  Chapters{Keep: true},
		Metadata:  Metadata{Keep: true},
		Output:    Output{Suffix: "-goencode"},
	}
}
