package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MediaInfo is what ffprobe found in a file.
type MediaInfo struct {
	Format     string // "matroska,webm"
	FormatName string // "Matroska / WebM"
	Duration   time.Duration
	Size       int64
	BitRate    int64
	Streams    []Stream
	Chapters   int
	Title      string
}

// Stream is one stream of a media file.
type Stream struct {
	Index       int // absolute index in the file, for -map 0:N
	Kind        MediaKind
	Codec       string // "hevc"
	CodecLong   string
	Profile     string
	Width       int
	Height      int
	FrameRate   float64
	PixelFormat string
	BitDepth    int
	Channels    int
	Layout      string
	SampleRate  int
	BitRate     int64
	Language    string
	Title       string
	Default     bool
	Forced      bool
	AttachedPic bool // cover art stored as a video stream
	HDR         bool
}

// TextSubtitle reports whether a subtitle stream is text (convertible to
// another text format) rather than bitmaps (PGS, VobSub, DVB).
func (s Stream) TextSubtitle() bool {
	switch s.Codec {
	case "hdmv_pgs_subtitle", "dvd_subtitle", "dvb_subtitle", "xsub", "dvb_teletext":
		return false
	}
	return true
}

// StreamsOf lists the streams of one kind, leaving out cover art.
func (m *MediaInfo) StreamsOf(k MediaKind) []Stream {
	var list []Stream
	for _, s := range m.Streams {
		if s.Kind == k && !s.AttachedPic {
			list = append(list, s)
		}
	}
	return list
}

// Video returns the main video stream.
func (m *MediaInfo) Video() (Stream, bool) {
	v := m.StreamsOf(KindVideo)
	if len(v) == 0 {
		return Stream{}, false
	}
	return v[0], true
}

// Probe runs ffprobe on a file.
func Probe(ctx context.Context, b Binaries, path string) (*MediaInfo, error) {
	out, err := command(ctx, b.FFprobe, "-v", "error", "-print_format", "json",
		"-show_format", "-show_streams", "-show_chapters", path).Output()
	if err != nil {
		if ee, ok := err.(interface{ ExitCode() int }); ok && ee.ExitCode() != 0 {
			return nil, fmt.Errorf("not a media file ffmpeg can read")
		}
		return nil, err
	}
	return parseProbe(out)
}

type probeJSON struct {
	Format struct {
		FormatName     string            `json:"format_name"`
		FormatLongName string            `json:"format_long_name"`
		Duration       string            `json:"duration"`
		Size           string            `json:"size"`
		BitRate        string            `json:"bit_rate"`
		Tags           map[string]string `json:"tags"`
	} `json:"format"`
	Streams []struct {
		Index         int               `json:"index"`
		CodecType     string            `json:"codec_type"`
		CodecName     string            `json:"codec_name"`
		CodecLongName string            `json:"codec_long_name"`
		Profile       string            `json:"profile"`
		Width         int               `json:"width"`
		Height        int               `json:"height"`
		AvgFrameRate  string            `json:"avg_frame_rate"`
		RFrameRate    string            `json:"r_frame_rate"`
		PixFmt        string            `json:"pix_fmt"`
		BitsPerRaw    string            `json:"bits_per_raw_sample"`
		Channels      int               `json:"channels"`
		ChannelLayout string            `json:"channel_layout"`
		SampleRate    string            `json:"sample_rate"`
		BitRate       string            `json:"bit_rate"`
		ColorTransfer string            `json:"color_transfer"`
		Tags          map[string]string `json:"tags"`
		Disposition   map[string]int    `json:"disposition"`
	} `json:"streams"`
	Chapters []json.RawMessage `json:"chapters"`
}

func parseProbe(data []byte) (*MediaInfo, error) {
	var p probeJSON
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	m := &MediaInfo{
		Format:     p.Format.FormatName,
		FormatName: p.Format.FormatLongName,
		Duration:   seconds(p.Format.Duration),
		Size:       atoi64(p.Format.Size),
		BitRate:    atoi64(p.Format.BitRate),
		Chapters:   len(p.Chapters),
		Title:      tag(p.Format.Tags, "title"),
	}
	for _, s := range p.Streams {
		st := Stream{
			Index:       s.Index,
			Codec:       s.CodecName,
			CodecLong:   s.CodecLongName,
			Profile:     s.Profile,
			Width:       s.Width,
			Height:      s.Height,
			PixelFormat: s.PixFmt,
			BitDepth:    int(atoi64(s.BitsPerRaw)),
			Channels:    s.Channels,
			Layout:      s.ChannelLayout,
			SampleRate:  int(atoi64(s.SampleRate)),
			BitRate:     atoi64(s.BitRate),
			Language:    tag(s.Tags, "language"),
			Title:       tag(s.Tags, "title"),
			Default:     s.Disposition["default"] == 1,
			Forced:      s.Disposition["forced"] == 1,
			AttachedPic: s.Disposition["attached_pic"] == 1,
			HDR:         s.ColorTransfer == "smpte2084" || s.ColorTransfer == "arib-std-b67",
		}
		switch s.CodecType {
		case "video":
			st.Kind = KindVideo
		case "audio":
			st.Kind = KindAudio
		case "subtitle":
			st.Kind = KindSubtitle
		default:
			continue
		}
		st.FrameRate = rate(s.AvgFrameRate)
		if st.FrameRate == 0 {
			st.FrameRate = rate(s.RFrameRate)
		}
		if st.BitDepth == 0 && strings.Contains(st.PixelFormat, "10") {
			st.BitDepth = 10
		}
		m.Streams = append(m.Streams, st)
	}
	return m, nil
}

func tag(tags map[string]string, key string) string {
	for k, v := range tags {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func seconds(s string) time.Duration {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return time.Duration(f * float64(time.Second))
}

func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func rate(s string) float64 {
	num, den, ok := strings.Cut(s, "/")
	if !ok {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	n, _ := strconv.ParseFloat(num, 64)
	d, _ := strconv.ParseFloat(den, 64)
	if d == 0 {
		return 0
	}
	return n / d
}
