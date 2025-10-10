package app

import "errors"

// --- FILE & FILE VERSION ---
var (
	ErrFileNotFound      = errors.New("file not found")
	ErrVersionNotFound   = errors.New("version not found")
	ErrCannotDeleteCurr  = errors.New("cannot delete current version")
	ErrVersionProcessing = errors.New("cannot delete file, some versions are processing")
)

// --- USER ---
var (
	ErrUserNotFound = errors.New("user not found")
)

// --- SESSION ---
var (
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidSession  = errors.New("invalid session")
	ErrInvalidExpiry   = errors.New("expiresAt must be in the future")
)

// --- PUBLIC LINK ---
var (
	ErrPublicLinkNotFound = errors.New("public link not found")
	ErrInvalidExpiryTime  = errors.New("expiresAt must be in the future")
)

// --- MAGIC LINK ---
var (
	ErrMagicLinkNotFound = errors.New("magic link not found")
	ErrInvalidMagicLink  = errors.New("invalid or expired magic link")
)
