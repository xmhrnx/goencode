// Package platform covers what differs between desktops and phones: where
// GoEncode keeps its files, and opening results in other apps.
package platform

import (
	"encoding/json"
	"os"
	"path/filepath"

	gioapp "gioui.org/app"

	"github.com/xmhrnx/goencode/internal/settings"
)

// Prefs are the app preferences, saved between runs.
type Prefs struct {
	FFmpegPath  string            `json:"ffmpegPath"`
	Concurrency int               `json:"concurrency"`
	Theme       string            `json:"theme"` // "system", "dark", "light"
	Advanced    bool              `json:"advanced"`
	Notify      bool              `json:"notify"`
	Preset      string            `json:"preset"`
	Settings    settings.Settings `json:"settings"`
}

// DefaultPrefs are the preferences of a first run.
func DefaultPrefs() Prefs {
	return Prefs{Concurrency: 1, Theme: "system", Notify: true, Preset: "universal", Settings: settings.Default()}
}

// DataDir is GoEncode's own folder: preferences, imported files and, on
// phones, finished files before they're shared.
func DataDir() string {
	dir, err := gioapp.DataDir()
	if err != nil || dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "GoEncode")
}

func prefsPath() string { return filepath.Join(DataDir(), "prefs.json") }

// LoadPrefs reads the saved preferences, or the defaults.
func LoadPrefs() Prefs {
	p := DefaultPrefs()
	data, err := os.ReadFile(prefsPath())
	if err != nil {
		return p
	}
	if json.Unmarshal(data, &p) != nil {
		return DefaultPrefs()
	}
	if p.Concurrency < 1 {
		p.Concurrency = 1
	}
	return p
}

// SavePrefs writes the preferences.
func SavePrefs(p Prefs) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(DataDir(), 0o755); err != nil {
		return err
	}
	tmp := prefsPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, prefsPath())
}

// ImportDir holds files copied in from a phone's picker.
func ImportDir() string { return filepath.Join(DataDir(), "imports") }

// OutputDir holds finished files on platforms where the source folder
// isn't writable.
func OutputDir() string { return filepath.Join(DataDir(), "output") }

// PreviewDir holds preview encodes.
func PreviewDir() string { return filepath.Join(os.TempDir(), "goencode-preview") }
