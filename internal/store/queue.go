package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/jobs"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/settings"
)

// Start queues every ready file (and failed or canceled ones, to retry)
// and starts encoding them with the current recipe.
func (s *Store) Start() {
	if !s.Ready() || s.Running {
		return
	}
	n := 0
	for _, it := range s.Items {
		switch it.Status {
		case jobs.Ready, jobs.Failed, jobs.Canceled, jobs.Done:
			if it.Status == jobs.Done && !s.restartDone() {
				continue
			}
			it.Status, it.Err, it.Log, it.Progress = jobs.Queued, "", nil, ffmpeg.Progress{}
			it.Started, it.Finished, it.OutSize = time.Time{}, time.Time{}, 0
			n++
		}
	}
	if n == 0 {
		s.Toast(ToastInfo, "Add some files first")
		return
	}
	s.Running = true
	if release, err := s.host.KeepAlive("Encoding " + plural(n, "file")); err == nil {
		s.release = release
	}
	s.pump()
}

// restartDone reports whether finished files are encoded again: only when
// nothing else is waiting, so pressing Start after a batch redoes it.
func (s *Store) restartDone() bool {
	for _, it := range s.Items {
		if it.Status == jobs.Ready || it.Status == jobs.Failed || it.Status == jobs.Canceled {
			return false
		}
	}
	return true
}

// Stop cancels every queued and running encode.
func (s *Store) Stop() {
	for _, it := range s.Items {
		if it.Status == jobs.Queued {
			it.Status = jobs.Canceled
		}
	}
	for _, cancel := range s.cancels {
		cancel()
	}
}

// Cancel stops one file.
func (s *Store) Cancel(id int) {
	it := s.Item(id)
	if it == nil {
		return
	}
	if it.Status == jobs.Queued {
		it.Status = jobs.Canceled
		s.pump()
	}
	if cancel, ok := s.cancels[id]; ok {
		cancel()
	}
}

// pump starts queued files up to the concurrency limit and ends the batch
// when nothing is left.
func (s *Store) pump() {
	limit := max(1, s.Prefs.Concurrency)
	for _, it := range s.Items {
		if len(s.cancels) >= limit {
			return
		}
		if it.Status == jobs.Queued {
			s.run(it)
		}
	}
	if len(s.cancels) == 0 && s.Running {
		s.finishBatch()
	}
}

func (s *Store) taken(path string) bool {
	for _, it := range s.Items {
		if it.Status.Active() && it.Output == path {
			return true
		}
	}
	return false
}

func (s *Store) run(it *jobs.Item) {
	fallback := ""
	if it.Imported {
		fallback = platform.OutputDir()
		os.MkdirAll(fallback, 0o755)
	}
	it.Output = ""
	name := it.Path
	if it.Disc != nil {
		// A disc's result is named after it and goes beside its folder,
		// or, for the disc itself, which can't be written to, with
		// results from phones.
		parent := filepath.Dir(it.Path)
		if parent == it.Path || it.Disc.Root == filepath.VolumeName(it.Path)+string(filepath.Separator) {
			parent = platform.OutputDir()
			os.MkdirAll(parent, 0o755)
		}
		name = filepath.Join(parent, fileName.Replace(it.Disc.Name)+".disc")
	}
	out := settings.OutputPath(name, s.Settings, fallback, s.taken)
	input, inArgs := it.Source()
	args, _, dur, err := settings.Args(settings.Job{Input: input, InputArgs: inArgs, Info: it.Info, Output: out}, s.Settings, s.Caps)
	if err != nil {
		it.Status, it.Err = jobs.Failed, err.Error()
		return
	}
	if dir := s.Settings.Output.Dir; dir != "" {
		os.MkdirAll(dir, 0o755)
	}
	ctx, cancel := context.WithCancel(s.ctx)
	s.cancels[it.ID] = cancel
	it.Status, it.Output, it.Started = jobs.Running, out, time.Now()
	bins := s.Bins

	go func() {
		var last time.Time
		err := ffmpeg.Run(ctx, bins, args, dur, func(p ffmpeg.Progress) {
			if time.Since(last) < 150*time.Millisecond && p.Fraction < 1 {
				return
			}
			last = time.Now()
			s.host.Post(func() { it.Progress = p })
		})
		size := int64(0)
		if fi, statErr := os.Stat(out); statErr == nil {
			size = fi.Size()
		}
		if err != nil {
			os.Remove(out) // a partial file is useless
		}
		s.host.Post(func() {
			cancel()
			delete(s.cancels, it.ID)
			it.Finished = time.Now()
			var runErr *ffmpeg.RunError
			switch {
			case err == nil:
				it.Status, it.OutSize, it.Progress.Fraction = jobs.Done, size, 1
			case ffmpeg.IsCanceled(err):
				it.Status = jobs.Canceled
			case errors.As(err, &runErr):
				it.Status, it.Err, it.Log = jobs.Failed, runErr.Summary(), runErr.Log
			default:
				it.Status, it.Err = jobs.Failed, err.Error()
			}
			s.pump()
		})
	}()
}

func (s *Store) finishBatch() {
	s.Running = false
	if s.release != nil {
		s.release()
		s.release = nil
	}
	done, failed := 0, 0
	var saved int64
	for _, it := range s.Items {
		switch it.Status {
		case jobs.Done:
			done++
			saved += it.Size - it.OutSize
		case jobs.Failed:
			failed++
		}
	}
	if done+failed == 0 {
		s.Toast(ToastInfo, "Encoding stopped")
		return
	}
	msg := plural(done, "file") + " encoded"
	if failed > 0 {
		msg += fmt.Sprintf(", %d failed", failed)
	}
	if saved > 0 {
		msg += " · " + FormatBytes(saved) + " saved"
	}
	kind := ToastSuccess
	if failed > 0 {
		kind = ToastWarning
	}
	s.Toast(kind, msg)
	if s.Prefs.Notify && done+failed > 0 {
		s.host.Notify("GoEncode finished", msg)
	}
}

// Progress is the batch's overall progress, weighted by duration.
func (s *Store) Progress() (fraction float64, running, queued int) {
	var total, done float64
	for _, it := range s.Items {
		if it.Info == nil {
			continue
		}
		d := it.Info.Duration.Seconds()
		switch it.Status {
		case jobs.Running:
			running++
			total += d
			done += d * it.Progress.Fraction
		case jobs.Queued:
			queued++
			total += d
		case jobs.Done:
			if s.Running {
				total += d
				done += d
			}
		}
	}
	if total == 0 {
		return 0, running, queued
	}
	return done / total, running, queued
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// FormatBytes prints a size: 1.4 GB, 312 MB.
func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// fileName makes a disc's title safe as a file name: "Film: Part 2".
var fileName = strings.NewReplacer(": ", " - ", ":", "-", "/", "-", "\\", "-", "?", "", "*", "", "\"", "'", "<", "", ">", "", "|", "-")
