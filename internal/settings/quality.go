package settings

import (
	"math"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// The simple mode shows one "smaller file ↔ better quality" slider for the
// video. These map its 0..1 position onto the current rate control's useful
// range, whichever direction the encoder's scale runs.

func qualityRange(rc *catalog.RateControl) (lo, hi float64) {
	if rc.Bitrate {
		return math.Max(rc.Min, rc.Default/4), math.Min(rc.Max, rc.Default*3)
	}
	span := (rc.Max - rc.Min) / 4
	return math.Max(rc.Min, rc.Default-span), math.Min(rc.Max, rc.Default+span)
}

// QualityFraction returns where the video quality sits on the simple
// slider, 0 being the smallest file and 1 the best quality. ok is false when
// the encoder has no quality to choose.
func QualityFraction(s Settings, caps *ffmpeg.Capabilities) (f float64, ok bool) {
	if s.Video.Mode != ModeEncode || s.Video.Encoder == "" {
		return 0, false
	}
	rc := ProfileOf(caps, s.Video.Encoder, ffmpeg.KindVideo).RateControl(s.Video.RateControl)
	if rc == nil || rc.None {
		return 0, false
	}
	lo, hi := qualityRange(rc)
	if rc.Bitrate {
		f = (math.Log(float64(s.Video.Bitrate)) - math.Log(lo)) / (math.Log(hi) - math.Log(lo))
	} else {
		f = (s.Video.Quality - lo) / (hi - lo)
		if rc.LowerIsBetter {
			f = 1 - f
		}
	}
	return math.Max(0, math.Min(1, f)), true
}

// SetQualityFraction moves the video quality to a slider position.
func SetQualityFraction(s *Settings, caps *ffmpeg.Capabilities, f float64) {
	rc := ProfileOf(caps, s.Video.Encoder, ffmpeg.KindVideo).RateControl(s.Video.RateControl)
	if rc == nil || rc.None {
		return
	}
	lo, hi := qualityRange(rc)
	if rc.Bitrate {
		v := math.Exp(math.Log(lo) + f*(math.Log(hi)-math.Log(lo)))
		s.Video.Bitrate = int(math.Round(v/rc.Step) * rc.Step)
		return
	}
	if rc.LowerIsBetter {
		f = 1 - f
	}
	v := lo + f*(hi-lo)
	step := rc.Step
	if step == 0 {
		step = 1
	}
	s.Video.Quality = math.Round(v/step) * step
}
