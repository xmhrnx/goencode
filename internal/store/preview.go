package store

import (
	"context"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/settings"
)

// PreviewState is where a preview encode is.
type PreviewState int

const (
	PreviewIdle PreviewState = iota
	PreviewRunning
	PreviewDone
	PreviewFailed
)

// Preview is a short test encode of part of one file, to judge the
// recipe's quality, speed and size before running the batch.
type Preview struct {
	ItemID   int
	Start    time.Duration
	Length   time.Duration
	State    PreviewState
	Progress ffmpeg.Progress
	Err      string
	Output   string
	Size     int64
	Took     time.Duration

	// Projected size of the whole file at the preview's bitrate.
	Projected int64
	// Source and encoded frames from the middle of the clip.
	Before, After image.Image

	cancel context.CancelFunc
	gen    int
}

// StartPreview encodes Length of the item from Start with the current
// recipe, then extracts the same frame from the source and the result.
func (s *Store) StartPreview(itemID int, start, length time.Duration) {
	it := s.Item(itemID)
	if it == nil || it.Info == nil || !s.Ready() {
		return
	}
	s.CancelPreview()
	if start+length > it.Info.Duration {
		start = max(0, it.Info.Duration-length)
	}
	os.MkdirAll(platform.PreviewDir(), 0o755)
	c := catalog.ContainerByID(s.Settings.Container)
	out := filepath.Join(platform.PreviewDir(), fmt.Sprintf("preview-%d-%d.%s", it.ID, time.Now().UnixNano(), c.Ext))
	input, inArgs := it.Source()
	args, _, dur, err := settings.Args(settings.Job{Input: input, InputArgs: inArgs, Info: it.Info, Output: out,
		Window: settings.Window{Start: start, Duration: length}}, s.Settings, s.Caps)
	if old := s.Preview.Output; old != "" {
		os.Remove(old)
	}
	gen := s.Preview.gen + 1
	s.Preview = Preview{ItemID: itemID, Start: start, Length: length, State: PreviewRunning, gen: gen}
	if err != nil {
		s.Preview.State, s.Preview.Err = PreviewFailed, err.Error()
		return
	}
	ctx, cancel := context.WithCancel(s.ctx)
	s.Preview.cancel = cancel
	bins := s.Bins
	_, hasVideo := it.Info.Video()
	keepVideo := hasVideo && c.Holds(ffmpeg.KindVideo) && s.Settings.Video.Mode != settings.ModeDrop && c.ID != "gif"
	full := it.Info.Duration
	began := time.Now()

	go func() {
		err := ffmpeg.Run(ctx, bins, args, dur, func(p ffmpeg.Progress) {
			s.host.Post(func() {
				if s.Preview.gen == gen && s.Preview.State == PreviewRunning {
					s.Preview.Progress = p
				}
			})
		})
		took := time.Since(began)
		var size int64
		if fi, e := os.Stat(out); e == nil {
			size = fi.Size()
		}
		var before, after image.Image
		if err == nil && keepVideo {
			mid := dur / 2
			before, _ = ffmpeg.Frame(ctx, bins, input, start+mid, 960, inArgs...)
			after, _ = ffmpeg.Frame(ctx, bins, out, mid, 960)
		}
		s.host.Post(func() {
			if s.Preview.gen != gen || s.Preview.cancel == nil {
				return // replaced or canceled meanwhile
			}
			p := &s.Preview
			p.cancel = nil
			var runErr *ffmpeg.RunError
			switch {
			case err == nil:
				p.State, p.Output, p.Size, p.Took = PreviewDone, out, size, took
				p.Before, p.After = before, after
				p.Progress.Fraction = 1
				if dur > 0 {
					p.Projected = int64(float64(size) * float64(full) / float64(dur))
				}
			case ffmpeg.IsCanceled(err):
				p.State = PreviewIdle
			case errors.As(err, &runErr):
				p.State, p.Err = PreviewFailed, runErr.Summary()
			default:
				p.State, p.Err = PreviewFailed, err.Error()
			}
		})
	}()
}

// CancelPreview stops a running preview.
func (s *Store) CancelPreview() {
	if s.Preview.cancel != nil {
		s.Preview.cancel()
		s.Preview.cancel = nil
		s.Preview.State = PreviewIdle
	}
}
