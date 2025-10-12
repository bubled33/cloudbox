package file

import (
	"errors"
	"strings"
)

// FileName VO для имени файла
type FileName struct {
	value string
}

// NewFileName создаёт новый FileName с валидацией
func NewFileName(raw string) (FileName, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return FileName{}, errors.New("file name cannot be empty")
	}
	if len(name) > 255 { // ограничение длины имени файла
		return FileName{}, errors.New("file name cannot exceed 255 characters")
	}
	return FileName{value: name}, nil
}

// String возвращает строковое значение имени файла
func (f FileName) String() string {
	return f.value
}
