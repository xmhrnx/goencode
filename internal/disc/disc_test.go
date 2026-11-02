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
		it := make([]byte, 2+20)
		binary.BigEndian.PutUint16(it, 20)
		copy(it[2:], c+"M2TS")
		binary.BigEndian.PutUint32(it[14:], 900000)
		binary.BigEndian.PutUint32(it[18:], uint32(900000+secs*45000))
		items = append(items, it...)
	}
	return append(b, items...)
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
