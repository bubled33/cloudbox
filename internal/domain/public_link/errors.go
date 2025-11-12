package public_link

import "errors"

var (
	ErrNotFound          = errors.New("public link not found")
	ErrInvalidExpiryTime = errors.New("expiresAt must be in the future")
)
