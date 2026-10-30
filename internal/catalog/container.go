// Package catalog is GoEncode's knowledge about containers, codecs and
// encoder options: which codecs a container accepts, how each encoder's
// quality is controlled, and the explanations shown when hovering an option.
package catalog

import (
	"path"
	"strings"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// Container is an output format.
type Container struct {
	ID          string
	Label       string
	Ext         string
	Muxer       string // the -f value
	AudioOnly   bool
	Description string
	// Accepted codec IDs per kind. "*" accepts any codec of that kind;
	// patterns use path.Match syntax ("pcm_*"). Empty accepts none.
	Video, Audio, Subtitle []string
	Chapters               bool
	// Preferred codecs, best first, when the current one isn't accepted.
	PreferVideo, PreferAudio, PreferSubtitle []string
	// Extra output arguments.
	Args []string
}

// Accepts reports whether the container can hold a codec of a kind.
func (c *Container) Accepts(kind ffmpeg.MediaKind, codec string) bool {
	var list []string
	switch kind {
	case ffmpeg.KindVideo:
		list = c.Video
	case ffmpeg.KindAudio:
		list = c.Audio
	case ffmpeg.KindSubtitle:
		list = c.Subtitle
	}
	for _, pat := range list {
		if pat == "*" {
			return !rejectedEverywhere(kind, codec)
		}
		if ok, _ := path.Match(pat, codec); ok {
			return true
		}
	}
	return false
}

// Holds reports whether the container has room for a kind of stream at all.
func (c *Container) Holds(kind ffmpeg.MediaKind) bool {
	switch kind {
	case ffmpeg.KindVideo:
		return len(c.Video) > 0
	case ffmpeg.KindAudio:
		return len(c.Audio) > 0
	case ffmpeg.KindSubtitle:
		return len(c.Subtitle) > 0
	}
	return false
}

// rejectedEverywhere lists codecs even permissive containers can't store:
// image-sequence formats and pseudo codecs that only make sense on their own.
func rejectedEverywhere(kind ffmpeg.MediaKind, codec string) bool {
	switch codec {
	case "a64_multi", "a64_multi5", "wrapped_avframe", "vnull", "anull", "dvb_teletext", "ttml":
		return true
	}
	return false
}

// ContainerByID finds a container, falling back to Matroska.
func ContainerByID(id string) *Container {
	for i := range Containers {
		if Containers[i].ID == id {
			return &Containers[i]
		}
	}
	return &Containers[0]
}

// ContainerForExt finds the container of a file extension, or nil.
func ContainerForExt(ext string) *Container {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for i := range Containers {
		if Containers[i].Ext == ext {
			return &Containers[i]
		}
	}
	return nil
}
