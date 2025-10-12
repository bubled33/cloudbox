package public_link

import "errors"

var (
	ErrPublicLinkNotFound = errors.New("public link not found")
	ErrInvalidExpiryTime  = errors.New("expiresAt must be in the future")
)
