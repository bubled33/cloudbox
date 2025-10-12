package magic_link

import "errors"

// --- MAGIC LINK ---
var (
	ErrNotFound  = errors.New("magic link not found")
	ErrMagicLink = errors.New("invalid or expired magic link")
)
