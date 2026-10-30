package settings

import (
	"strings"
	"testing"
	"time"

	"github.com/xmhrnx/goencode/internal/ffmpeg"
)

const codecTable = ` -------
 DEV.LS h264                 H.264 / AVC (decoders: h264) (encoders: libx264 libx264rgb h264_nvenc)
 DEV.L. hevc                 H.265 / HEVC (decoders: hevc) (encoders: libx265 hevc_nvenc hevc_vaapi)
 DEV.L. av1                  AV1 (decoders: libdav1d) (encoders: libsvtav1 libaom-av1)
 DEV.L. vp9                  VP9 (encoders: libvpx-vp9)
 DEVIL. prores               Apple ProRes (encoders: prores prores_aw prores_ks)
 DEV.L. gif                  GIF
 DEA.L. aac                  AAC (encoders: aac)
 DEA.L. opus                 Opus (decoders: opus libopus) (encoders: opus libopus)
 DEA.L. mp3                  MP3 (decoders: mp3float mp3) (encoders: libmp3lame)
 DEA..S flac                 FLAC
 DEA.L. ac3                  ATSC A/52A (AC-3)
 D.A.L. dts                  DCA (DTS Coherent Acoustics) (decoders: dca)
 DEA..S pcm_s16le            PCM signed 16-bit little-endian
 DES... subrip               SubRip subtitle (decoders: srt subrip) (encoders: srt subrip)
 DES... ass                  ASS (encoders: ssa ass)
 DES... mov_text             MOV text
 DES... webvtt               WebVTT subtitle
 D.S... hdmv_pgs_subtitle    HDMV Presentation Graphic Stream subtitles (decoders: pgssub)
`

func testCaps() *ffmpeg.Capabilities {
	return ffmpeg.NewCapabilities(ffmpeg.Binaries{FFmpeg: "ffmpeg"}, ffmpeg.ParseCodecs(codecTable), "hevc_nvenc")
}

func testInfo() *ffmpeg.MediaInfo {
	return &ffmpeg.MediaInfo{
		Duration: 90 * time.Second,
		Chapters: 3,
		Streams: []ffmpeg.Stream{
			{Index: 0, Kind: ffmpeg.KindVideo, Codec: "hevc", Width: 3840, Height: 2160},
			{Index: 1, Kind: ffmpeg.KindAudio, Codec: "pcm_s24le", Channels: 6, Language: "eng"},
			{Index: 2, Kind: ffmpeg.KindAudio, Codec: "aac", Channels: 2, Language: "jpn"},
			{Index: 3, Kind: ffmpeg.KindSubtitle, Codec: "subrip", Language: "eng"},
			{Index: 4, Kind: ffmpeg.KindSubtitle, Codec: "hdmv_pgs_subtitle", Language: "eng", Forced: true},
			{Index: 5, Kind: ffmpeg.KindVideo, Codec: "mjpeg", AttachedPic: true},
		},
	}
}

func joined(a []string) string { return strings.Join(a, " ") }

func TestCodecsRespectContainer(t *testing.T) {
	caps := testCaps()
	s := Default()
	s.Container = "webm"
	got := Normalize(s, caps)
	if got.Video.Codec != "vp9" || got.Video.Encoder != "libvpx-vp9" {
		t.Errorf("webm video = %s/%s, want vp9/libvpx-vp9", got.Video.Codec, got.Video.Encoder)
	}
	if got.Audio.Codec != "opus" || got.Audio.Encoder != "libopus" {
		t.Errorf("webm audio = %s/%s, want opus/libopus", got.Audio.Codec, got.Audio.Encoder)
	}
}

func TestAudioOnlyContainerDropsVideo(t *testing.T) {
	caps := testCaps()
	s := Default()
	s.Container = "flac"
	got := Normalize(s, caps)
	if got.Video.Mode != ModeDrop {
		t.Errorf("video mode = %s, want drop", got.Video.Mode)
	}
	if got.Audio.Codec != "flac" || got.Audio.RateControl != "lossless" {
		t.Errorf("audio = %s %s", got.Audio.Codec, got.Audio.RateControl)
	}
}

func TestEncoderSwitchResetsRateControl(t *testing.T) {
	caps := testCaps()
	s := Default()
	s.Video.Codec, s.Video.Encoder = "hevc", "hevc_nvenc"
	got := Normalize(s, caps)
	if got.Video.RateControl != "cq" || got.Video.Quality != 24 {
		t.Errorf("nvenc rate = %s %v, want cq 24", got.Video.RateControl, got.Video.Quality)
	}
}

func TestUntestedHardwareIsHidden(t *testing.T) {
	caps := testCaps()
	for _, e := range Encoders(caps, "h264") {
		if e == "h264_nvenc" {
			t.Fatal("h264_nvenc offered without passing its test")
		}
		if e == "libx264rgb" {
			t.Fatal("libx264rgb shouldn't be offered")
		}
	}
	if got := Encoders(caps, "hevc"); len(got) != 2 || got[1] != "hevc_nvenc" {
		t.Errorf("hevc encoders = %v", got)
	}
}

func TestRemuxToMP4(t *testing.T) {
	caps := testCaps()
	s := Default()
	s.Video.Mode, s.Audio.Mode = ModeCopy, ModeCopy
	s = Normalize(s, caps)
	args, plan, _, err := Args(Job{Input: "in.mkv", Info: testInfo(), Output: "out.mp4"}, s, caps)
	if err != nil {
		t.Fatal(err)
	}
	a := joined(args)
	for _, want := range []string{
		"-map 0:0 -c:v:0 copy -tag:v:0 hvc1",
		"-map 0:1 -c:a:0 aac -b:a:0 160k", // PCM can't go in MP4: converted
		"-map 0:2 -c:a:1 copy",
		"-map 0:3 -c:s:0 mov_text", // SRT converted to MP4's text format
		"-movflags +faststart",
		"-f mp4 -y out.mp4",
	} {
		if !strings.Contains(a, want) {
			t.Errorf("args lack %q:\n%s", want, a)
		}
	}
	if strings.Contains(a, "0:4") || strings.Contains(a, "0:5") {
		t.Errorf("PGS subtitles or cover art mapped into MP4:\n%s", a)
	}
	if w := plan.Warnings(); len(w) != 3 {
		t.Errorf("warnings = %q, want PCM, SRT and PGS notes", w)
	}
}

func TestEncodeWithOptionsAndScaling(t *testing.T) {
	caps := testCaps()
	s := Default()
	s.Container = "mkv"
	s.Video.Height = 1080
	s.Video.PixelFormat = "yuv420p10le"
	s.SetVideoOption("tune", "film")
	s.Audio.Tracks, s.Audio.Languages = TracksLanguages, "jpn"
	s.Audio.Channels = 2
	s.Chapters.Keep = false
	s = Normalize(s, caps)
	args, _, _, err := Args(Job{Input: "in.mkv", Info: testInfo(), Output: "out.mkv"}, s, caps)
	if err != nil {
		t.Fatal(err)
	}
	a := joined(args)
	for _, want := range []string{
		"-vf scale=-2:1080:flags=lanczos",
		"-c:v:0 libx264 -crf:v:0 21 -preset:v:0 medium -tune:v:0 film -pix_fmt:v:0 yuv420p10le",
		"-map 0:2 -c:a:0",
		"-map_chapters -1",
	} {
		if !strings.Contains(a, want) {
			t.Errorf("args lack %q:\n%s", want, a)
		}
	}
	if strings.Contains(a, "-map 0:1 ") {
		t.Errorf("English audio kept despite language filter:\n%s", a)
	}
	if strings.Contains(a, "-ac:a:0") {
		t.Errorf("stereo source downmixed to stereo:\n%s", a)
	}
}

func TestBurnImageSubtitlesAndPreviewWindow(t *testing.T) {
	caps := testCaps()
	s := Default()
	s.Subtitles.Mode = SubsBurn
	s = Normalize(s, caps)
	args, plan, dur, err := Args(Job{Input: "in.mkv", Info: testInfo(), Output: "p.mp4",
		Window: Window{Start: 30 * time.Second, Duration: 10 * time.Second}}, s, caps)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Burn == nil || plan.Burn.Index != 4 {
		t.Fatalf("burn track = %+v, want the forced PGS track", plan.Burn)
	}
	a := joined(args)
	if !strings.HasPrefix(a, "-ss 30.000 -i in.mkv -t 10.000") {
		t.Errorf("window args wrong:\n%s", a)
	}
	if !strings.Contains(a, "-filter_complex [0:0][0:4]overlay=eof_action=pass,scale") && !strings.Contains(a, "[0:0][0:4]overlay=eof_action=pass[vout]") {
		t.Errorf("overlay missing:\n%s", a)
	}
	if dur != 10*time.Second {
		t.Errorf("duration = %v", dur)
	}
}

func TestQualityFractionRoundTrip(t *testing.T) {
	caps := testCaps()
	s := Normalize(Default(), caps)
	SetQualityFraction(&s, caps, 1)
	best := s.Video.Quality
	SetQualityFraction(&s, caps, 0)
	if s.Video.Quality <= best {
		t.Errorf("x264: smallest file CRF %v should be above best-quality CRF %v", s.Video.Quality, best)
	}
	SetQualityFraction(&s, caps, 0.5)
	if f, ok := QualityFraction(s, caps); !ok || f < 0.4 || f > 0.6 {
		t.Errorf("fraction = %v", f)
	}
}

func TestPresetsNormalize(t *testing.T) {
	caps := testCaps()
	for _, p := range AvailablePresets(caps) {
		s := Normalize(p.Build(caps), caps)
		if _, _, _, err := Args(Job{Input: "in.mkv", Info: testInfo(), Output: "out"}, s, caps); err != nil {
			t.Errorf("preset %s: %v", p.ID, err)
		}
	}
}
