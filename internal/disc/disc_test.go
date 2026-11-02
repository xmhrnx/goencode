package disc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mpls builds a playlist that plays clips, each for secs seconds.
func mpls(clips []string, secs int) []byte {
	b := make([]byte, 40)
	copy(b, "MPLS0200")
	binary.BigEndian.PutUint32(b[8:], 40)
	items := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.BigEndian.PutUint16(items[6:], uint16(len(clips)))
	for _, c := range clips {
		it := make([]byte, 34)
		copy(it[2:], c+"M2TS")
		binary.BigEndian.PutUint32(it[14:], 900000)
		binary.BigEndian.PutUint32(it[18:], uint32(900000+secs*45000))
		it = append(it, stn()...)
		binary.BigEndian.PutUint16(it, uint16(len(it)-2))
		items = append(items, it...)
	}
	return append(b, items...)
}

// stn is a stream table with a video, English and French audio, and a
// German subtitle.
func stn() []byte {
	t := make([]byte, 16)
	t[4], t[5], t[6] = 1, 2, 1
	entry := func(pid uint16, attr ...byte) {
		e := make([]byte, 10)
		e[0], e[1] = 9, 1
		binary.BigEndian.PutUint16(e[2:], pid)
		a := make([]byte, 6)
		a[0] = 5
		copy(a[1:], attr)
		t = append(append(t, e...), a...)
	}
	entry(0x1011, 0x1B, 0x61)
	entry(0x1100, 0x83, 0x61, 'e', 'n', 'g')
	entry(0x1101, 0x81, 0x31, 'f', 'r', 'a')
	entry(0x1200, 0x90, 'g', 'e', 'r')
	binary.BigEndian.PutUint16(t, uint16(len(t)-2))
	return t
}

func TestMPLSLanguages(t *testing.T) {
	pl, err := ParseMPLS(mpls([]string{"00010", "00011"}, 60))
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]string{0x1100: "eng", 0x1101: "fra", 0x1200: "ger"}
	if len(pl.Languages) != len(want) {
		t.Fatalf("languages = %v", pl.Languages)
	}
	for pid, l := range want {
		if pl.Languages[pid] != l {
			t.Errorf("PID %#x = %q, want %q", pid, pl.Languages[pid], l)
		}
	}
	if pl.Duration != 2*time.Minute || len(pl.Clips) != 2 {
		t.Errorf("playlist = %+v", pl)
	}
}

// ifo builds a title set IFO with program chains of the given lengths
// (whole minutes, at 25 fps).
func ifo(minutes ...int) []byte {
	b := make([]byte, 2048*2)
	copy(b, "DVDVIDEO-VTS")
	binary.BigEndian.PutUint32(b[0xCC:], 1)
	base := 2048
	binary.BigEndian.PutUint16(b[base:], uint16(len(minutes)))
	for i, m := range minutes {
		off := 8 + len(minutes)*8 + i*16
		binary.BigEndian.PutUint32(b[base+8+i*8+4:], uint32(off))
		t := b[base+off+4:]
		t[0], t[1], t[2], t[3] = byte(m/60/10<<4|m/60%10), byte(m%60/10<<4|m%60%10), 0x30, 0x40|0x12
	}
	return b
}

func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBluRay(t *testing.T) {
	root := filepath.Join(t.TempDir(), "MY_FILM")
	write(t, filepath.Join(root, "BDMV", "index.bdmv"), []byte("INDX0200"))
	write(t, filepath.Join(root, "BDMV", "PLAYLIST", "00000.mpls"), mpls([]string{"00001"}, 30))
	write(t, filepath.Join(root, "BDMV", "PLAYLIST", "00800.mpls"), mpls([]string{"00010", "00011"}, 3600))
	write(t, filepath.Join(root, "BDMV", "PLAYLIST", "00900.mpls"), mpls([]string{"00099"}, 9000)) // clip missing

	// A menu's background, looped for 5 hours.
	loop := make([]string, 300)
	for i := range loop {
		loop[i] = "00001"
	}
	write(t, filepath.Join(root, "BDMV", "PLAYLIST", "00149.mpls"), mpls(loop, 60))
	for _, c := range []string{"00001", "00010", "00011"} {
		write(t, filepath.Join(root, "BDMV", "STREAM", c+".m2ts"), make([]byte, 100))
	}
	write(t, filepath.Join(root, "BDMV", "META", "DL", "bdmt_eng.xml"), []byte(`<di:title><di:name>My Film: Part 2</di:name></di:title>`))

	for _, p := range []string{root, filepath.Join(root, "BDMV")} {
		if got := Root(p); got != root {
			t.Errorf("Root(%s) = %q, want the disc folder", p, got)
		}
	}
	if Root(filepath.Join(root, "BDMV", "STREAM")) != "" {
		t.Error("STREAM taken for a disc")
	}
	ti, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if ti.Kind != BluRay || ti.Source != "00800.mpls" || len(ti.Files) != 2 || ti.Duration != 2*time.Hour || ti.Size != 200 {
		t.Errorf("title = %+v", ti)
	}
	if ti.Name != "My Film: Part 2" {
		t.Errorf("name = %q", ti.Name)
	}
	in, args, err := ti.Input(t.TempDir())
	if err != nil || filepath.Ext(in) != ".ffconcat" || args[1] != "concat" {
		t.Errorf("input = %q %v %v", in, args, err)
	}
}

func TestDVD(t *testing.T) {
	root := filepath.Join(t.TempDir(), "SERIES")
	ts := filepath.Join(root, "VIDEO_TS")
	write(t, filepath.Join(ts, "VIDEO_TS.IFO"), []byte("x"))
	write(t, filepath.Join(ts, "VTS_01_0.VOB"), make([]byte, 50))
	write(t, filepath.Join(ts, "VTS_01_1.VOB"), make([]byte, 50))
	write(t, filepath.Join(ts, "VTS_02_0.VOB"), make([]byte, 10))
	write(t, filepath.Join(ts, "VTS_02_2.VOB"), make([]byte, 80))
	write(t, filepath.Join(ts, "VTS_02_1.VOB"), make([]byte, 80))
	write(t, filepath.Join(ts, "VTS_02_0.IFO"), ifo(5, 95, 1))

	if got := Root(ts); got != root {
		t.Errorf("Root(VIDEO_TS) = %q", got)
	}
	ti, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if ti.Kind != DVD || ti.Source != "VTS_02" || len(ti.Files) != 2 || filepath.Base(ti.Files[0]) != "VTS_02_1.VOB" || ti.Size != 160 {
		t.Errorf("title = %+v", ti)
	}
	if want := 95*time.Minute + 30*time.Second + 12*time.Second/25; ti.Duration != want {
		t.Errorf("duration = %v, want %v", ti.Duration, want)
	}
	if ti.Name != "SERIES" {
		t.Errorf("name = %q", ti.Name)
	}
	in, _, _ := ti.Input("")
	if in[:7] != "concat:" {
		t.Errorf("input = %q", in)
	}
}

func TestNotADisc(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a.m2ts"), []byte("x"))
	if Root(dir) != "" {
		t.Error("a plain folder taken for a disc")
	}
}
