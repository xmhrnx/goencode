package views

import (
	"fmt"
	"strconv"

	"github.com/xmhrnx/goencode/internal/catalog"
	"github.com/xmhrnx/goencode/internal/ffmpeg"
	"github.com/xmhrnx/goencode/internal/settings"
	"github.com/xmhrnx/goencode/internal/store"
	"github.com/xmhrnx/goencode/ui/components"
)

// Helpers shared by the settings views: the label/value lists their
// dropdowns show, built from what the container and ffmpeg allow.

func codecChoices(s *store.Store, kind ffmpeg.MediaKind) (labels, values []string) {
	c := catalog.ContainerByID(s.Settings.Container)
	for _, id := range settings.Codecs(s.Caps, c, kind) {
		desc := ""
		if cd := s.Caps.Codecs[id]; cd != nil {
			desc = cd.Description
		}
		labels = append(labels, catalog.Codec(id, desc).Label)
		values = append(values, id)
	}
	return
}

func codecHelp(s *store.Store, id string) string {
	desc := ""
	if cd := s.Caps.Codecs[id]; cd != nil {
		desc = cd.Description
	}
	c := catalog.ContainerByID(s.Settings.Container)
	return catalog.Codec(id, desc).Help + " Only codecs " + c.Label + " can hold are listed."
}

func encoderChoices(s *store.Store, codec string, kind ffmpeg.MediaKind) (labels, values []string) {
	for _, e := range settings.Encoders(s.Caps, codec) {
		p := settings.ProfileOf(s.Caps, e, kind)
		label := p.Label
		if p.Hardware != "" {
			label += "  ·  GPU"
		}
		labels = append(labels, label)
		values = append(values, e)
	}
	return
}

func rateChoices(p *catalog.Profile) (labels, values []string) {
	for _, rc := range p.RateControls {
		labels = append(labels, rc.Label)
		values = append(values, rc.ID)
	}
	return
}

// rateSlider describes the quality slider of a rate control.
func rateSlider(rc *catalog.RateControl) components.SliderSpec {
	spec := components.SliderSpec{Min: rc.Min, Max: rc.Max, Step: rc.Step}
	if rc.LowerIsBetter {
		spec.Reverse = true
		spec.LeftLabel, spec.RightLabel = "Smaller file", "Higher quality"
	} else {
		spec.LeftLabel, spec.RightLabel = "Smaller file", "Higher quality"
	}
	flag := rc.Flag
	spec.Format = func(v float64) string {
		label := "Q"
		switch flag {
		case "crf":
			label = "CRF"
		case "cq":
			label = "CQ"
		case "qp", "qp_i":
			label = "QP"
		}
		return label + " " + strconv.FormatFloat(v, 'f', -1, 64)
	}
	return spec
}

func markFor(rc *catalog.RateControl, v float64) string {
	best := ""
	dist := 1e9
	for _, m := range rc.Marks {
		if d := abs(m.Value - v); d <= 1.5 && d < dist {
			best, dist = m.Label, d
		}
	}
	if best == "" && v == rc.Default {
		return "recommended"
	}
	return best
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func kbps(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.1f Mb/s", v/1000)
	}
	return fmt.Sprintf("%.0f kb/s", v)
}

var resolutions = []struct {
	label  string
	height int
}{
	{"Same as source", 0}, {"2160p · 4K", 2160}, {"1440p · QHD", 1440}, {"1080p · Full HD", 1080},
	{"720p · HD", 720}, {"576p", 576}, {"480p", 480}, {"360p", 360}, {"240p", 240},
}

var frameRates = []struct {
	label string
	rate  float64
}{
	{"Same as source", 0}, {"23.976", 24000.0 / 1001}, {"24", 24}, {"25", 25}, {"29.97", 30000.0 / 1001},
	{"30", 30}, {"50", 50}, {"59.94", 60000.0 / 1001}, {"60", 60}, {"15", 15}, {"12", 12}, {"10", 10},
}

var channelChoices = []struct {
	label string
	n     int
}{
	{"Same as source", 0}, {"Mono", 1}, {"Stereo", 2}, {"5.1 surround", 6}, {"7.1 surround", 8},
}
