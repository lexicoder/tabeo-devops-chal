package booking

import (
	"errors"
)

var (
	ErrInvalidUUID          = errors.New("invalid uuid")
	ErrMissingDestination   = errors.New("destination does not exist")
	ErrLaunchPadUnavailable = errors.New("launchpad is unavailable")
)
