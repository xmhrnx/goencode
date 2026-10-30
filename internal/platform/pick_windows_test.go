//go:build windows

package platform

import (
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestSplitPicked(t *testing.T) {
	enc := func(parts ...string) []uint16 {
		var b []uint16
		for _, p := range parts {
			w, _ := syscall.UTF16FromString(p)
			b = append(b, w...)
		}
		return append(b, 0, 0, 0)
	}
	one := filepath.Join(`C:\media`, "a.mkv")
	if got := splitPicked(enc(one)); !reflect.DeepEqual(got, []string{one}) {
		t.Errorf("one file: %q", got)
	}
	want := []string{filepath.Join(`C:\media`, "a.mkv"), filepath.Join(`C:\media`, "b.mp4")}
	if got := splitPicked(enc(`C:\media`, "a.mkv", "b.mp4")); !reflect.DeepEqual(got, want) {
		t.Errorf("two files: %q", got)
	}
	if got := splitPicked(make([]uint16, 8)); len(got) != 0 {
		t.Errorf("empty: %q", got)
	}
}
