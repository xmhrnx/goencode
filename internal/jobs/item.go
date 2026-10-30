// Package jobs holds the files in the batch and their encoding state.
package jobs

import (
	"image"
	"time"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// Status is where a file is in its life.
type Status int

const (
	Probing Status = iota
	Ready
	Invalid // ffprobe couldn't read it
	Queued
	Running
	Done
	Failed
	Canceled
)

// String returns the status for display.
func (s Status) String() string {
	return [...]string{"Analyzing", "Ready", "Unreadable", "Queued", "Encoding", "Done", "Failed", "Canceled"}[s]
}

// Active reports whether the file is waiting for or in an encode.
func (s Status) Active() bool { return s == Queued || s == Running }

// Item is one file in the batch.
type Item struct {
	ID       int
	Path     string
	Name     string
	Size     int64
	Imported bool // copied into app storage from a phone's picker

	Info  *ffmpeg.MediaInfo
	Thumb image.Image

	Status   Status
	Progress ffmpeg.Progress
	Output   string
	OutSize  int64
	Err      string
	Log      []string
	Started  time.Time
	Finished time.Time
}

// Elapsed is how long the last encode took or has taken so far.
func (it *Item) Elapsed() time.Duration {
	if it.Started.IsZero() {
		return 0
	}
	if it.Finished.IsZero() {
		return time.Since(it.Started)
	}
	return it.Finished.Sub(it.Started)
}

// Ratio is the output size relative to the input, once done.
func (it *Item) Ratio() float64 {
	if it.Size == 0 || it.OutSize == 0 {
		return 0
	}
	return float64(it.OutSize) / float64(it.Size)
}
