package user

import (
	"fmt"
	"net/mail"
	"strings"
)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(raw))
	fmt.Println(raw, addr, err)
	if err != nil {
		return Email{}, ErrInvalidEmailFormat
	}
	return Email{value: strings.ToLower(addr.Address)}, nil
}

func (e Email) String() string {
	return e.value
}

type DisplayName struct {
	value string
}

func NewDisplayName(raw string) (DisplayName, error) {
	name := strings.TrimSpace(raw)
	if len(name) < 2 || len(name) > 50 {
		return DisplayName{}, ErrInvalidDisplayNameSize
	}
	return DisplayName{value: name}, nil
}

func (d DisplayName) String() string {
	return d.value
}
