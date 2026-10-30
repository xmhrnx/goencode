package components

import (
	"github.com/mehran9675/giode"

	"github.com/xmhrnx/goencode/internal/jobs"
)

// StatusPill shows where a file is: analyzing, ready, encoding, done...
func StatusPill(s jobs.Status) giode.Element {
	switch s {
	case jobs.Probing:
		return Pill(s.String(), "", "clock")
	case jobs.Ready:
		return Pill(s.String(), "info", "check")
	case jobs.Queued:
		return Pill(s.String(), "", "clock")
	case jobs.Running:
		return Pill(s.String(), "primary", "zap")
	case jobs.Done:
		return Pill(s.String(), "success", "check")
	case jobs.Canceled:
		return Pill(s.String(), "warning", "stop")
	default:
		return Pill(s.String(), "danger", "alert")
	}
}
