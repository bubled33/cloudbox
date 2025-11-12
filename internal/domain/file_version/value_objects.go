package file_version

import (
	"errors"
	"fmt"
)

type S3Key struct {
	value string
}

func NewS3Key(value string) (S3Key, error) {
	if value == "" {
		return S3Key{}, errors.New("S3 key cannot be empty")
	}
	return S3Key{value: value}, nil
}

func (k S3Key) String() string {
	return k.value
}

type MimeType struct {
	value string
}

func NewMimeType(value string) (MimeType, error) {
	if value == "" {
		return MimeType{}, errors.New("MIME type cannot be empty")
	}
	return MimeType{value: value}, nil
}

func (m MimeType) String() string {
	return m.value
}

type FileSize struct {
	value uint64
}

const MaxFileSize = 10 * 1024 * 1024 * 1024 // 10 GB

func NewFileSize(value uint64) (FileSize, error) {
	if value > MaxFileSize {
		return FileSize{}, fmt.Errorf("file size exceeds 10GB limit")
	}
	return FileSize{value: value}, nil
}

func (s FileSize) Uint64() uint64 {
	return s.value
}

type FileVersionNum struct {
	value int
}

func NewFileVersionNum(value int) (FileVersionNum, error) {
	if value < 1 {
		return FileVersionNum{}, errors.New("file version number must be >= 1")
	}
	return FileVersionNum{value: value}, nil
}

func (v FileVersionNum) Int() int {
	return v.value
}

func (v FileVersionNum) Equal(o FileVersionNum) bool {
	return v.value == o.value
}

type FileStatus string

const (
	FileStatusUploaded   FileStatus = "uploaded"
	FileStatusProcessing FileStatus = "processing"
	FileStatusReady      FileStatus = "ready"
	FileStatusFailed     FileStatus = "failed"
)

func NewFileStatus(status string) (FileStatus, error) {
	switch FileStatus(status) {
	case FileStatusUploaded, FileStatusProcessing, FileStatusReady, FileStatusFailed:
		return FileStatus(status), nil
	default:
		return "", fmt.Errorf("invalid file status: %s", status)
	}
}

func (s FileStatus) String() string {
	return string(s)
}

func (s FileStatus) Equal(other FileStatus) bool {
	return s == other
}
