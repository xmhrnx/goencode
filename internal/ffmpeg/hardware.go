package ffmpeg

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"
)

// HardwareFamily describes a family of hardware encoders by the suffix
// ffmpeg gives them.
type HardwareFamily struct {
	Suffix string // "_nvenc"
	Name   string // "NVIDIA NVENC"
	// Args go before the input for encoders that need a device opened.
	Args []string
	// Upload is appended to the video filter chain to move frames to the
	// device, for encoders that don't accept frames in system memory.
	Upload string
}

// HardwareFamilies are the hardware encoder families GoEncode knows.
var HardwareFamilies = []HardwareFamily{
	{Suffix: "_nvenc", Name: "NVIDIA NVENC"},
	{Suffix: "_qsv", Name: "Intel Quick Sync"},
	{Suffix: "_amf", Name: "AMD AMF"},
	{Suffix: "_videotoolbox", Name: "Apple VideoToolbox"},
	{Suffix: "_vaapi", Name: "VA-API", Args: []string{"-vaapi_device", "/dev/dri/renderD128"}, Upload: "format=nv12,hwupload"},
	{Suffix: "_mf", Name: "Windows Media Foundation"},
	{Suffix: "_mediacodec", Name: "Android MediaCodec"},
	{Suffix: "_v4l2m2m", Name: "V4L2 M2M"},
	{Suffix: "_d3d12va", Name: "Direct3D 12"},
}

// HardwareOf returns the family of a hardware encoder.
func HardwareOf(encoder string) (HardwareFamily, bool) {
	for _, f := range HardwareFamilies {
		if strings.HasSuffix(encoder, f.Suffix) {
			return f, true
		}
	}
	return HardwareFamily{}, false
}

// TestHardware tries every hardware video encoder ffmpeg lists with a tiny
// test encode: an encoder being built in doesn't mean this computer has the
// hardware or driver for it. It runs once; WorkingHardware reads the result.
func (c *Capabilities) TestHardware(ctx context.Context) {
	var candidates []string
	for enc, codec := range c.EncoderCodec {
		if _, ok := HardwareOf(enc); ok && c.Codecs[codec].Kind == KindVideo {
			candidates = append(candidates, enc)
		}
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)
	for _, enc := range candidates {
		wg.Add(1)
		go func(enc string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			ok := c.testEncoder(ctx, enc)
			c.mu.Lock()
			c.working[enc] = ok
			c.mu.Unlock()
		}(enc)
	}
	wg.Wait()
	c.mu.Lock()
	c.tested = true
	c.mu.Unlock()
}

func (c *Capabilities) testEncoder(ctx context.Context, enc string) bool {
	if strings.HasSuffix(enc, "_mediacodec") && runtime.GOOS != "android" {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	fam, _ := HardwareOf(enc)
	args := []string{"-hide_banner", "-nostdin", "-loglevel", "error"}
	args = append(args, fam.Args...)
	args = append(args, "-f", "lavfi", "-i", "color=black:s=320x240:r=30:d=0.2")
	if fam.Upload != "" {
		args = append(args, "-vf", fam.Upload)
	} else {
		args = append(args, "-pix_fmt", "yuv420p")
	}
	args = append(args, "-frames:v", "3", "-c:v", enc, "-f", "null", "-")
	return command(ctx, c.Binaries.FFmpeg, args...).Run() == nil
}

// HardwareTested reports whether TestHardware has finished.
func (c *Capabilities) HardwareTested() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tested
}

// EncoderUsable reports whether an encoder can be offered: software
// encoders always, hardware ones once they passed the test encode.
func (c *Capabilities) EncoderUsable(enc string) bool {
	if !c.HasEncoder(enc) {
		return false
	}
	if _, hw := HardwareOf(enc); !hw {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.working[enc]
}

// WorkingHardware lists the hardware families with at least one encoder
// that passed the test.
func (c *Capabilities) WorkingHardware() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	seen := map[string]bool{}
	var names []string
	for _, f := range HardwareFamilies {
		for enc, ok := range c.working {
			if ok && strings.HasSuffix(enc, f.Suffix) && !seen[f.Name] {
				seen[f.Name] = true
				names = append(names, f.Name)
			}
		}
	}
	return names
}
