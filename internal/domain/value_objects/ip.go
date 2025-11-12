package value_objects

import (
	"database/sql/driver"
	"fmt"
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

// 👇 Добавляем реализацию интерфейса sql.Scanner
func (i *IP) Scan(value interface{}) error {
	if value == nil {
		i.value = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		i.value = net.ParseIP(v)
	case []byte:
		i.value = net.ParseIP(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into IP", value)
	}

	if i.value == nil {
		return domainerrors.ErrInvalidIP
	}

	return nil
}

// 👇 Добавляем реализацию интерфейса driver.Valuer
func (i IP) Value() (driver.Value, error) {
	if i.value == nil {
		return nil, nil
	}
	return i.value.String(), nil
}
