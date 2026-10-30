package ffmpeg

import (
	"context"
	"testing"
)

func TestParseEncoderHelp(t *testing.T) {
	out := `Encoder libx264 [libx264 H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10]:
    General capabilities: dr1 delay threads
    Supported pixel formats: yuv420p yuv420p10le
libx264 AVOptions:
  -preset            <string>     E..V....... Set the encoding preset (cf. x264 --fullhelp) (default "medium")
  -crf               <float>      E..V....... Select the quality for constant quality mode (from -1 to FLT_MAX) (default -1)
  -aq-mode           <int>        E..V....... AQ method (from -1 to INT_MAX) (default -1)
     none            0            E..V.......
     variance        1            E..V....... Variance AQ (complexity mask)
`
	h := parseEncoderHelp("libx264", out)
	if h.Hardware || len(h.PixelFormats) != 2 || len(h.Options) != 3 {
		t.Fatalf("parsed %+v", h)
	}
	if o := h.Options[0]; o.Name != "preset" || o.Default != "medium" || o.Description != "Set the encoding preset (cf. x264 --fullhelp)" {
		t.Errorf("preset = %+v", o)
	}
	if o := h.Options[1]; o.Min != "-1" || o.Max != "FLT_MAX" {
		t.Errorf("crf range = %+v", o)
	}
	if c := h.Options[2].Constants; len(c) != 2 || c[1].Name != "variance" || c[1].Description != "Variance AQ (complexity mask)" {
		t.Errorf("aq-mode constants = %+v", c)
	}
}

func TestDetectRealFFmpeg(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	ctx := context.Background()
	b, err := Locate(ctx, "")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	c, err := Detect(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	if c.Codecs["h264"] == nil || c.Codecs["aac"] == nil {
		t.Fatal("h264 or aac missing from codec table")
	}
	c.TestHardware(ctx)
	t.Logf("ffmpeg %s, %d encodable codecs, working hardware: %v", b.Version, len(c.Codecs), c.WorkingHardware())
	h := c.EncoderHelp(ctx, "aac")
	if len(h.SampleRates) == 0 || len(h.Options) == 0 {
		t.Errorf("aac help = %+v", h)
	}
}
