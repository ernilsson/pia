package profile

import "errors"

var (
	ErrProfileNotFound    = errors.New("profile not found")
	ErrInvalidProfileLine = errors.New("invalid profile line")
)
