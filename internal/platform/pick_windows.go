//go:build windows

package platform

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// GioUI's explorer mishandles a single file picked in a multi-select dialog
// (Windows then returns its full path instead of "folder\0name") and reports
// it as canceled, so the Windows picker is called directly.

var (
	comdlg32             = syscall.NewLazyDLL("comdlg32.dll")
	getOpenFileName      = comdlg32.NewProc("GetOpenFileNameW")
	commDlgExtendedError = comdlg32.NewProc("CommDlgExtendedError")
)

const (
	ofnAllowMultiSelect = 0x00000200
	ofnExplorer         = 0x00080000
	ofnFileMustExist    = 0x00001000
	ofnNoChangeDir      = 0x00000008
	ofnPathMustExist    = 0x00000800
	pickBufferLen       = 1 << 16
)

// openFileName is OPENFILENAMEW.
type openFileName struct {
	structSize      uint32
	owner           uintptr
	instance        uintptr
	filter          *uint16
	customFilter    *uint16
	maxCustomFilter uint32
	filterIndex     uint32
	file            *uint16
	maxFile         uint32
	fileTitle       *uint16
	maxFileTitle    uint32
	initialDir      *uint16
	title           *uint16
	flags           uint32
	fileOffset      uint16
	fileExtension   uint16
	defExt          *uint16
	custData        uintptr
	hook            uintptr
	templateName    *uint16
	reserved        uintptr
	reserved2       uint32
	flagsEx         uint32
}

// NativePickFiles shows the Windows open dialog for several files of the
// given extensions ("mkv", no dot) and returns their paths. ok is false on
// platforms where the app's own picker should be used instead.
func NativePickFiles(title, filterName string, exts []string) (paths []string, ok bool, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	patterns := make([]string, len(exts))
	for i, e := range exts {
		patterns[i] = "*." + strings.TrimPrefix(e, ".")
	}
	// Pairs of "name\0pattern\0", ended by an extra NUL.
	filter := filterName + "\x00" + strings.Join(patterns, ";") + "\x00All files\x00*.*\x00\x00"
	filterW := utf16.Encode([]rune(filter))
	titleW, _ := syscall.UTF16PtrFromString(title)
	buf := make([]uint16, pickBufferLen)

	ofn := openFileName{
		filter:  &filterW[0],
		file:    &buf[0],
		maxFile: uint32(len(buf)),
		title:   titleW,
		flags:   ofnAllowMultiSelect | ofnExplorer | ofnFileMustExist | ofnPathMustExist | ofnNoChangeDir,
	}
	ofn.structSize = uint32(unsafe.Sizeof(ofn))

	if r, _, _ := getOpenFileName.Call(uintptr(unsafe.Pointer(&ofn))); r == 0 {
		if code, _, _ := commDlgExtendedError.Call(); code != 0 {
			return nil, true, fmt.Errorf("file dialog error %#x", code)
		}
		return nil, true, ErrPickCanceled
	}
	return splitPicked(buf), true, nil
}

// splitPicked reads the dialog's result: one full path, or a folder followed
// by file names, each ended by NUL and the list by an empty string.
func splitPicked(buf []uint16) []string {
	var parts []string
	for start := 0; start < len(buf) && buf[start] != 0; {
		end := start
		for end < len(buf) && buf[end] != 0 {
			end++
		}
		parts = append(parts, syscall.UTF16ToString(buf[start:end]))
		start = end + 1
	}
	if len(parts) <= 1 {
		return parts
	}
	paths := make([]string, len(parts)-1)
	for i, name := range parts[1:] {
		paths[i] = filepath.Join(parts[0], name)
	}
	return paths
}
