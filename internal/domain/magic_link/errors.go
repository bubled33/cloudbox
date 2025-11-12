package magic_link

import "errors"

var (
	ErrNotFound  = errors.New("magic link not found")
	ErrInvalid   = errors.New("magic link is invalid")
	ErrMagicLink = errors.New("invalid or expired magic link")
)
