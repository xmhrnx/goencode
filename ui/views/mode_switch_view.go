package views

import (
	"github.com/xmhrnx/giode"

	"github.com/xmhrnx/goencode/ui/components"
)

// ModeSwitch flips between the simple settings (presets and one quality
// slider) and the advanced ones (every option of every codec). key keeps
// the header's and the phone layout's switches apart.
func ModeSwitch(env *Env, key string) giode.Element {
	s := env.Store
	mode := "simple"
	if s.Prefs.Advanced {
		mode = "advanced"
	}
	return components.Segmented("mode-"+key, []string{"Simple", "Advanced"}, []string{"simple", "advanced"}, []string{"sparkle", "tune"}, mode,
		func(v string) { s.SetAdvanced(v == "advanced") })
}
