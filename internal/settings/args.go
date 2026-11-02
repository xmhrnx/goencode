package settings

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

// Window limits an encode to part of the file, for previews.
type Window struct {
	Start    time.Duration
	Duration time.Duration
}

// Job is everything needed to turn a file into ffmpeg arguments.
type Job struct {
	Input     string
	InputArgs []string // options for the input, such as the concat demuxer
	Info   *ffmpeg.MediaInfo
	Output string
	Window Window
}

// Args builds the ffmpeg arguments for a job (after the global flags that
// ffmpeg.Run adds) and returns the duration ffmpeg will write, for progress.
func Args(j Job, s Settings, caps *ffmpeg.Capabilities) ([]string, *Plan, time.Duration, error) {
	if j.Info == nil {
		return nil, nil, 0, fmt.Errorf("the file hasn't been analyzed yet")
	}
	p := MakePlan(j.Info, s, caps)
	if p.Video == nil && len(nonDropped(p.Audio)) == 0 {
		return nil, p, 0, fmt.Errorf("nothing to write: no video or audio track fits %s", p.Container.Label)
	}

	var a []string
	encVideo := p.Video != nil && p.Video.Action != ActCopy
	var hw ffmpeg.HardwareFamily
	isHW := false
	if encVideo {
		hw, isHW = ffmpeg.HardwareOf(p.Video.Encoder)
		if isHW {
			a = append(a, hw.Args...)
		}
		if s.Video.HWDecode {
			a = append(a, "-hwaccel", "auto")
		}
	}
	if j.Window.Start > 0 {
		a = append(a, "-ss", secs(j.Window.Start))
	}
	a = append(a, j.InputArgs...)
	a = append(a, "-i", j.Input)
	duration := j.Info.Duration
	if j.Window.Duration > 0 {
		a = append(a, "-t", secs(j.Window.Duration))
		if rest := j.Info.Duration - j.Window.Start; rest > 0 && rest < j.Window.Duration {
			duration = rest
		} else {
			duration = j.Window.Duration
		}
	}

	// Video.
	if p.Video != nil {
		v := p.Video
		if !encVideo {
			a = append(a, "-map", fmt.Sprintf("0:%d", v.Stream.Index), "-c:v:0", "copy")
		} else {
			filters := videoFilters(j, s, p, hw, isHW)
			if p.Burn != nil && !p.Burn.TextSubtitle() {
				graph := fmt.Sprintf("[0:%d][0:%d]overlay=eof_action=pass", v.Stream.Index, p.Burn.Index)
				if len(filters) > 0 {
					graph += "," + strings.Join(filters, ",")
				}
				a = append(a, "-filter_complex", graph+"[vout]", "-map", "[vout]")
			} else {
				a = append(a, "-map", fmt.Sprintf("0:%d", v.Stream.Index))
				if len(filters) > 0 {
					a = append(a, "-vf", strings.Join(filters, ","))
				}
			}
			a = append(a, "-c:v:0", v.Encoder)
			prof := ProfileOf(caps, v.Encoder, ffmpeg.KindVideo)
			rc := prof.RateControl(s.Video.RateControl)
			if v.Action == ActConvert && s.Video.Mode != ModeEncode {
				// A forced conversion uses the encoder's defaults.
				rc = prof.RateControl("")
				a = append(a, rateArgs(rc, "v:0", rcDefault(rc), int(rcDefault(rc)), 0)...)
			} else {
				a = append(a, rateArgs(rc, "v:0", s.Video.Quality, s.Video.Bitrate, s.Video.MaxBitrate)...)
				a = append(a, optionArgs(s.Video.Options[v.Encoder], "v:0")...)
				if s.Video.PixelFormat != "" && hw.Upload == "" {
					a = append(a, "-pix_fmt:v:0", s.Video.PixelFormat)
				}
				if s.Video.FrameRate > 0 && v.Encoder != "gif" {
					a = append(a, "-r:v:0", strconv.FormatFloat(s.Video.FrameRate, 'f', -1, 64))
				}
			}
		}
		if v.Codec == "hevc" && (p.Container.ID == "mp4" || p.Container.ID == "mov") {
			a = append(a, "-tag:v:0", "hvc1") // lets Apple devices play HEVC in MP4
		}
	}

	// Audio.
	ai := 0
	for _, sp := range p.Audio {
		if sp.Action == ActDrop {
			continue
		}
		spec := fmt.Sprintf("a:%d", ai)
		a = append(a, "-map", fmt.Sprintf("0:%d", sp.Stream.Index))
		if sp.Action == ActCopy {
			a = append(a, "-c:"+spec, "copy")
			ai++
			continue
		}
		a = append(a, "-c:"+spec, sp.Encoder)
		prof := ProfileOf(caps, sp.Encoder, ffmpeg.KindAudio)
		rc := prof.RateControl(s.Audio.RateControl)
		if sp.Action == ActConvert && s.Audio.Mode != ModeEncode {
			rc = prof.RateControl("")
			a = append(a, rateArgs(rc, spec, rcDefault(rc), int(rcDefault(rc)), 0)...)
		} else {
			a = append(a, rateArgs(rc, spec, s.Audio.Quality, s.Audio.Bitrate, 0)...)
			a = append(a, optionArgs(s.Audio.Options[sp.Encoder], spec)...)
			if s.Audio.Channels > 0 && (sp.Stream.Channels == 0 || s.Audio.Channels < sp.Stream.Channels) {
				a = append(a, "-ac:"+spec, strconv.Itoa(s.Audio.Channels))
			}
			if s.Audio.SampleRate > 0 {
				a = append(a, "-ar:"+spec, strconv.Itoa(s.Audio.SampleRate))
			}
			if s.Audio.Loudnorm {
				a = append(a, "-filter:"+spec, "loudnorm=I=-16:TP=-1.5:LRA=11")
			}
		}
		ai++
	}

	// Subtitles.
	si := 0
	for _, sp := range p.Subtitles {
		if sp.Action == ActDrop {
			continue
		}
		spec := fmt.Sprintf("s:%d", si)
		a = append(a, "-map", fmt.Sprintf("0:%d", sp.Stream.Index))
		if sp.Action == ActCopy {
			a = append(a, "-c:"+spec, "copy")
		} else {
			enc := DefaultEncoder(caps, sp.Codec)
			if enc == "" {
				enc = sp.Codec
			}
			a = append(a, "-c:"+spec, enc)
		}
		si++
	}

	if !p.Chapters {
		a = append(a, "-map_chapters", "-1")
	}
	if !s.Metadata.Keep {
		a = append(a, "-map_metadata", "-1")
	}
	a = append(a, p.Container.Args...)
	a = append(a, "-strict", "experimental", "-max_muxing_queue_size", "4096", "-f", p.Container.Muxer, "-y", j.Output)
	return a, p, duration, nil
}

func videoFilters(j Job, s Settings, p *Plan, hw ffmpeg.HardwareFamily, isHW bool) []string {
	var f []string
	forced := p.Video.Action == ActConvert && s.Video.Mode != ModeEncode
	if p.Burn != nil && p.Burn.TextSubtitle() {
		if j.Window.Start > 0 {
			// Input seeking restarts timestamps at zero; shift them back so
			// the subtitles line up, then restore.
			f = append(f, "setpts=PTS+"+secs(j.Window.Start)+"/TB")
		}
		f = append(f, fmt.Sprintf("subtitles=%s:si=%d", filterPath(j.Input), subtitleOrdinal(j.Info, p.Burn.Index)))
		if j.Window.Start > 0 {
			f = append(f, "setpts=PTS-STARTPTS")
		}
	}
	if !forced && s.Video.Height > 0 && p.Video.Stream.Height > s.Video.Height {
		f = append(f, fmt.Sprintf("scale=-2:%d:flags=lanczos", s.Video.Height))
	}
	if p.Video.Encoder == "gif" {
		if !forced && s.Video.FrameRate > 0 {
			f = append(f, "fps="+strconv.FormatFloat(s.Video.FrameRate, 'f', -1, 64))
		}
		f = append(f, "split[g0][g1];[g0]palettegen=stats_mode=diff[gp];[g1][gp]paletteuse=dither=sierra2_4a")
	}
	if isHW && hw.Upload != "" {
		f = append(f, hw.Upload)
	}
	return f
}

// subtitleOrdinal converts a stream's absolute index into its position
// among the subtitle streams, which the subtitles filter's si expects.
func subtitleOrdinal(info *ffmpeg.MediaInfo, index int) int {
	n := 0
	for _, s := range info.Streams {
		if s.Index == index {
			return n
		}
		if s.Kind == ffmpeg.KindSubtitle {
			n++
		}
	}
	return 0
}

// filterPath escapes a file path for use inside a filtergraph argument.
func filterPath(p string) string {
	if runtime.GOOS == "windows" || strings.Contains(p, `\`) {
		p = strings.ReplaceAll(p, `\`, "/")
	}
	p = strings.NewReplacer(`'`, `\'`, ":", `\:`).Replace(p)
	return "'" + p + "'"
}

func rcDefault(rc *catalog.RateControl) float64 {
	if rc == nil {
		return 0
	}
	return rc.Default
}

func rateArgs(r *catalog.RateControl, spec string, q float64, br, maxbr int) []string {
	if r == nil {
		return nil
	}
	out := append([]string(nil), r.Args...)
	if r.None {
		return out
	}
	if r.Bitrate {
		out = append(out, "-b:"+spec, strconv.Itoa(br)+"k")
		if maxbr > 0 {
			out = append(out, "-maxrate:"+spec, strconv.Itoa(maxbr)+"k", "-bufsize:"+spec, strconv.Itoa(maxbr*2)+"k")
		}
		return out
	}
	val := strconv.FormatFloat(q, 'f', -1, 64)
	for _, flag := range append([]string{r.Flag}, r.AlsoFlags...) {
		if strings.Contains(flag, ":") {
			out = append(out, "-"+flag, val)
		} else {
			out = append(out, "-"+flag+":"+spec, val)
		}
	}
	return out
}

func optionArgs(opts map[string]string, spec string) []string {
	keys := make([]string, 0, len(opts))
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []string
	for _, k := range keys {
		if v := strings.TrimSpace(opts[k]); v != "" {
			out = append(out, "-"+k+":"+spec, v)
		}
	}
	return out
}

func nonDropped(list []StreamPlan) []StreamPlan {
	var out []StreamPlan
	for _, s := range list {
		if s.Action != ActDrop {
			out = append(out, s)
		}
	}
	return out
}

func secs(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 3, 64)
}
