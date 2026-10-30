package platform

import "errors"

// ErrPickCanceled means the user closed NativePickFiles without choosing.
var ErrPickCanceled = errors.New("canceled")
