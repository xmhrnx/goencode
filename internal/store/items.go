package store

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/xmhrnx/goencode/internal/disc"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/jobs"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/settings"
)

// AddPaths adds files (and the media files inside folders) to the batch.
func (s *Store) AddPaths(paths []string) {
	added := 0
	for _, p := range jobs.Expand(paths) {
		if s.has(p) {
			continue
		}
		s.addItem(p, false)
		added++
	}
	if added == 0 && len(paths) > 0 {
		s.Toast(ToastInfo, "Those files are already in the list")
	}
}

// PickedFile is a file chosen in a system picker.
type PickedFile struct {
	Name string
	Path string // empty on phones
	Body io.ReadCloser
}

// AddPicked adds files from a picker, copying them into app storage when
// the platform gives no path. It runs off the UI goroutine.
func (s *Store) AddPicked(files []PickedFile) {
	for _, f := range files {
		if f.Path != "" {
			f.Body.Close()
			path := f.Path
			s.host.Post(func() { s.AddPaths([]string{path}) })
			continue
		}
		path, err := platform.Import(f.Name, f.Body)
		f.Body.Close()
		if err != nil {
			s.host.Post(func() { s.Toast(ToastError, "Couldn't import "+f.Name+": "+err.Error()) })
			continue
		}
		s.host.Post(func() { s.addItem(path, true) })
	}
}

func (s *Store) has(path string) bool {
	for _, it := range s.Items {
		if it.Path == path {
			return true
		}
	}
	return false
}

func (s *Store) addItem(path string, imported bool) {
	s.nextID++
	it := &jobs.Item{ID: s.nextID, Path: path, Name: filepath.Base(path), Imported: imported, Status: jobs.Probing}
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		it.Size = fi.Size()
	}
	s.Items = append(s.Items, it)
	if s.Selected == 0 {
		s.Selected = it.ID
	}
	if s.Ready() {
		s.probe(it)
	}
}

// probe reads a file's streams and a thumbnail in the background, a few
// files at a time: a dropped folder can hold hundreds. A disc folder's
// main title is found first.
func (s *Store) probe(it *jobs.Item) {
	bins := s.Bins
	path := it.Path
	go func() {
		select {
		case s.probes <- struct{}{}:
			defer func() { <-s.probes }()
		case <-s.ctx.Done():
			return
		}
		var title *disc.Title
		input, inArgs := path, []string(nil)
		if fi, err := os.Stat(path); err == nil && fi.IsDir() {
			var err error
			if title, err = disc.Open(path); err == nil {
				input, inArgs, err = title.Input(platform.PreviewDir())
			}
			if err != nil {
				s.host.Post(func() { it.Status, it.Err = jobs.Invalid, err.Error() })
				return
			}
		}
		ctx, cancel := context.WithTimeout(s.ctx, 90*time.Second)
		defer cancel()
		info, err := ffmpeg.Probe(ctx, bins, input, inArgs...)
		if err == nil && title != nil && title.Duration > 0 {
			info.Duration = title.Duration
		}
		s.host.Post(func() {
			if title != nil {
				it.Disc, it.Input, it.InputArgs = title, input, inArgs
				it.Name, it.Size = title.Name+" ("+title.Kind.String()+")", title.Size
			}
			if err != nil {
				it.Status, it.Err = jobs.Invalid, err.Error()
				return
			}
			it.Info, it.Status = info, jobs.Ready
		})
		if err != nil {
			return
		}
		if _, ok := info.Video(); ok {
			at := info.Duration / 10
			if at > 30*time.Second {
				at = 30 * time.Second
			}
			if img, err := ffmpeg.Frame(ctx, bins, input, at, 320, inArgs...); err == nil {
				s.host.Post(func() { it.Thumb = img })
			}
		}
	}()
}

// Item finds an item by ID.
func (s *Store) Item(id int) *jobs.Item {
	for _, it := range s.Items {
		if it.ID == id {
			return it
		}
	}
	return nil
}

// SelectedItem returns the selected item, or the first, or nil.
func (s *Store) SelectedItem() *jobs.Item {
	if it := s.Item(s.Selected); it != nil {
		return it
	}
	if len(s.Items) > 0 {
		return s.Items[0]
	}
	return nil
}

// Remove takes an item out of the batch, stopping its encode.
func (s *Store) Remove(id int) {
	if cancel, ok := s.cancels[id]; ok {
		cancel()
	}
	for i, it := range s.Items {
		if it.ID == id {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			if it.Imported && !it.Status.Active() {
				os.Remove(it.Path)
			}
			break
		}
	}
	if s.Selected == id {
		s.Selected = 0
		if len(s.Items) > 0 {
			s.Selected = s.Items[0].ID
		}
	}
}

// ClearFinished removes done, failed and canceled files.
func (s *Store) ClearFinished() {
	var keep []*jobs.Item
	for _, it := range s.Items {
		switch it.Status {
		case jobs.Done, jobs.Failed, jobs.Canceled, jobs.Invalid:
			if s.Selected == it.ID {
				s.Selected = 0
			}
		default:
			keep = append(keep, it)
		}
	}
	s.Items = keep
}

// ClearAll empties the batch when nothing is encoding.
func (s *Store) ClearAll() {
	if s.Running {
		return
	}
	for _, it := range s.Items {
		if it.Imported {
			os.Remove(it.Path)
		}
	}
	s.Items, s.Selected = nil, 0
}

// Plan returns what the current recipe will do with an item.
func (s *Store) Plan(it *jobs.Item) *settings.Plan {
	if it.Info == nil || s.Caps == nil {
		return nil
	}
	return settings.MakePlan(it.Info, s.Settings, s.Caps)
}

// Totals sums the batch for the header.
func (s *Store) Totals() (count int, size int64, duration time.Duration) {
	for _, it := range s.Items {
		count++
		size += it.Size
		if it.Info != nil {
			duration += it.Info.Duration
		}
	}
	return
}
