package value_objects

import (
	"strings"

	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
)

type TokenHash struct {
	value string
}

func NewTokenHash(raw string) (TokenHash, error) {
	token := strings.TrimSpace(raw)
	if token == "" {
		return TokenHash{}, domainerrors.ErrInvaliTokenHash
	}
	return TokenHash{value: token}, nil
}

func (t TokenHash) String() string {
	return t.value
}
