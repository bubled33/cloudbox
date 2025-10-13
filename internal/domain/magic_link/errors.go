package magic_link

import "errors"

var (
	ErrNotFound  = errors.New("magic link not found")
	ErrMagicLink = errors.New("invalid or expired magic link")
)
