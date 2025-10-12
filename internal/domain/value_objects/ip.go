package value_objects

import (
	"net"

	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
)

type IP struct {
	value net.IP
}

func NewIP(ip net.IP) (IP, error) {
	if ip == nil || ip.IsUnspecified() {
		return IP{}, domainerrors.ErrInvalidIP
	}
	return IP{value: ip}, nil
}

func (i IP) NetIP() net.IP {
	return i.value
}

func (i IP) String() string {
	return i.value.String()
}
