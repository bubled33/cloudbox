package value_objects

import "strings"

type DeviceInfo struct {
	value string
}

func NewDeviceInfo(raw string) (DeviceInfo, error) {
	info := strings.TrimSpace(raw)
	if info == "" {
		info = "unknown"
	}
	return DeviceInfo{value: info}, nil
}

func (d DeviceInfo) String() string {
	return d.value
}
