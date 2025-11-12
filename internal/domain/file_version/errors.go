package file_version

import "errors"

var (
	ErrVersionNotFound   = errors.New("version not found")
	ErrCannotDeleteCurr  = errors.New("cannot delete current version")
	ErrVersionProcessing = errors.New("cannot delete file, some versions are processing")
	ErrVersionFailed     = errors.New("cannot delete file, some versions are processing")
)
