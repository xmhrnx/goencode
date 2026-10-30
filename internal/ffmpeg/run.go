package ffmpeg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Progress is a snapshot of a running encode.
type Progress struct {
	Fraction float64       // 0..1 of the expected duration
	Position time.Duration // output time written so far
	Speed    float64       // 2.5 = two and a half times real time
	FPS      float64
	Size     int64 // bytes written
	ETA      time.Duration
}

// Run runs ffmpeg with args (which must not include -progress; Run adds it)
// and reports progress against the expected output duration. On failure
// the error carries the last lines ffmpeg printed.
func Run(ctx context.Context, b Binaries, args []string, duration time.Duration, onProgress func(Progress)) error {
	full := append([]string{"-hide_banner", "-nostdin", "-nostats", "-progress", "pipe:1", "-stats_period", "0.4"}, args...)
	cmd := command(ctx, b.FFmpeg, full...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	tail := newTail(14)
	wg.Add(2)
	go func() {
		defer wg.Done()
		readProgress(stdout, duration, onProgress)
	}()
	go func() {
		defer wg.Done()
		sc := bufio.NewScanner(stderr)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			tail.add(sc.Text())
		}
	}()
	wg.Wait()
	err = cmd.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return &RunError{Err: err, Log: tail.lines()}
	}
	return nil
}

// RunError is a failed ffmpeg run with the end of its log.
type RunError struct {
	Err error
	Log []string
}

func (e *RunError) Error() string {
	if msg := e.Summary(); msg != "" {
		return msg
	}
	return e.Err.Error()
}

// Summary picks the line of the log most likely to explain the failure.
func (e *RunError) Summary() string {
	for i := len(e.Log) - 1; i >= 0; i-- {
		l := strings.ToLower(e.Log[i])
		if strings.Contains(l, "error") || strings.Contains(l, "invalid") || strings.Contains(l, "not supported") ||
			strings.Contains(l, "could not") || strings.Contains(l, "unable") || strings.Contains(l, "incompatible") {
			return strings.TrimSpace(e.Log[i])
		}
	}
	if len(e.Log) > 0 {
		return strings.TrimSpace(e.Log[len(e.Log)-1])
	}
	return ""
}

// IsCanceled reports whether err is a canceled run.
func IsCanceled(err error) bool {
	return errors.Is(err, context.Canceled)
}

func readProgress(r io.Reader, duration time.Duration, onProgress func(Progress)) {
	sc := bufio.NewScanner(r)
	var p Progress
	started := time.Now()
	for sc.Scan() {
		key, val, ok := strings.Cut(strings.TrimSpace(sc.Text()), "=")
		if !ok {
			continue
		}
		switch key {
		case "out_time_us", "out_time_ms": // both are microseconds in ffmpeg
			if us, err := strconv.ParseInt(val, 10, 64); err == nil && us > 0 {
				p.Position = time.Duration(us) * time.Microsecond
			}
		case "fps":
			p.FPS, _ = strconv.ParseFloat(val, 64)
		case "total_size":
			p.Size, _ = strconv.ParseInt(val, 10, 64)
		case "speed":
			p.Speed, _ = strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(val), "x"), 64)
		case "progress":
			if duration > 0 {
				p.Fraction = min(1, float64(p.Position)/float64(duration))
				if p.Fraction > 0.01 {
					elapsed := time.Since(started)
					p.ETA = time.Duration(float64(elapsed) * (1 - p.Fraction) / p.Fraction)
				}
			}
			if val == "end" {
				p.Fraction, p.ETA = 1, 0
			}
			if onProgress != nil {
				onProgress(p)
			}
		}
	}
}

type tail struct {
	mu  sync.Mutex
	max int
	buf []string
}

func newTail(n int) *tail { return &tail{max: n} }

func (t *tail) add(s string) {
	if strings.TrimSpace(s) == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, s)
	if len(t.buf) > t.max {
		t.buf = t.buf[len(t.buf)-t.max:]
	}
}

func (t *tail) lines() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.buf...)
}

// FormatDuration prints a duration as h:mm:ss or m:ss.
func FormatDuration(d time.Duration) string {
	s := int(d.Round(time.Second).Seconds())
	if s < 0 {
		s = 0
	}
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}
