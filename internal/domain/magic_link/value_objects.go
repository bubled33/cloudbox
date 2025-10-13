package magic_link

import "errors"

type Purpose string

const (
	PurposeLogin         Purpose = "login"
	PurposeResetPassword Purpose = "reset_password"
)

func NewPurpose(raw string) (Purpose, error) {
	switch raw {
	case string(PurposeLogin), string(PurposeResetPassword):
		return Purpose(raw), nil
	default:
		return "", errors.New("invalid purpose")
	}
}

func (p Purpose) String() string {
	return string(p)
}
