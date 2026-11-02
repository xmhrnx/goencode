// Package store is GoEncode's application state and the actions on it.
//
// The UI reads the Store directly while it builds each frame. Every field
// is changed on the UI goroutine only: background work (probing, encoding,
// previews) hands its results back through Host.Post.
package store

import (
	"context"
	"time"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/jobs"
	"github.com/xmhrnx/goencode/internal/platform"
	"github.com/xmhrnx/goencode/internal/settings"
)

// Host is what the store needs from the app.
type Host interface {
	// Post runs fn on the UI goroutine and redraws.
	Post(fn func())
	Notify(title, body string) error
	KeepAlive(reason string) (release func(), err error)
}

// EngineState is how far finding and testing ffmpeg has got.
type EngineState int

const (
	EngineLoading EngineState = iota
	EngineReady
	EngineMissing
	EngineUnsupported
)

// Store is the application state.
type Store struct {
	host Host
	ctx  context.Context

	Prefs    platform.Prefs
	Settings settings.Settings

	Engine    EngineState
	EngineErr string
	Bins      ffmpeg.Binaries
	Caps      *ffmpeg.Capabilities
	HWReady   bool // the hardware test has finished

	Items    []*jobs.Item
	Selected int // ID of the selected item, 0 for none
	nextID   int

	Running  bool
	cancels  map[int]context.CancelFunc
	probes   chan struct{} // limits how many files are analyzed at once
	release  func()
	batchEnd func()

	Preview Preview

	// Toast shows a message; set by the UI.
	Toast func(kind ToastKind, msg string)

	saveTimer *time.Timer
}

// ToastKind colors a toast.
type ToastKind int

const (
	ToastInfo ToastKind = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// New creates the store and starts looking for ffmpeg.
func New(ctx context.Context, host Host) *Store {
	s := &Store{host: host, ctx: ctx, cancels: map[int]context.CancelFunc{}, probes: make(chan struct{}, 2)}
	s.Prefs = platform.LoadPrefs()
	s.Settings = s.Prefs.Settings
	s.Toast = func(ToastKind, string) {}
	s.LoadEngine()
	return s
}

// LoadEngine finds ffmpeg (again, after the path in preferences changed),
// reads its capabilities and tests the hardware encoders.
func (s *Store) LoadEngine() {
	s.Engine, s.EngineErr, s.HWReady = EngineLoading, "", false
	override := s.Prefs.FFmpegPath
	go func() {
		bins, err := ffmpeg.Locate(s.ctx, override)
		var caps *ffmpeg.Capabilities
		if err == nil {
			caps, err = ffmpeg.Detect(s.ctx, bins)
		}
		s.host.Post(func() {
			switch {
			case err == ffmpeg.ErrUnsupported:
				s.Engine, s.EngineErr = EngineUnsupported, err.Error()
				return
			case err != nil:
				s.Engine, s.EngineErr = EngineMissing, err.Error()
				return
			}
			s.Engine, s.Bins, s.Caps = EngineReady, bins, caps
			s.Settings = settings.Normalize(s.Settings, caps)
			s.loadHelp()
			for _, it := range s.Items {
				if it.Status == jobs.Probing {
					s.probe(it)
				}
			}
			go func() {
				caps.TestHardware(s.ctx)
				s.host.Post(func() {
					s.HWReady = true
					s.Settings = settings.Normalize(s.Settings, caps)
				})
			}()
		})
	}()
}

// Ready reports whether ffmpeg is usable.
func (s *Store) Ready() bool { return s.Engine == EngineReady }

// SavePrefs stores the preferences soon, coalescing bursts of changes
// (a dragged slider) into one write.
func (s *Store) SavePrefs() {
	s.Prefs.Settings = s.Settings
	p := s.Prefs
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}
	s.saveTimer = time.AfterFunc(700*time.Millisecond, func() { platform.SavePrefs(p) })
}

// SetPrefs replaces the preferences, reloading ffmpeg when its path changed.
func (s *Store) SetPrefs(p platform.Prefs) {
	reload := p.FFmpegPath != s.Prefs.FFmpegPath
	s.Prefs = p
	s.SavePrefs()
	if reload && !s.Running {
		s.LoadEngine()
	}
}
