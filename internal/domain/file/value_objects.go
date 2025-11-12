package file

import (
	"errors"
	"strings"
)

type FileName struct {
	value string
}

func NewFileName(raw string) (FileName, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return FileName{}, errors.New("file name cannot be empty")
	}
	if len(name) > 255 {
		return FileName{}, errors.New("file name cannot exceed 255 characters")
	}
	return FileName{value: name}, nil
}

func (f FileName) String() string {
	return f.value
}
