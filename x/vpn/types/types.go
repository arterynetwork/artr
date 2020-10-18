package types

import "fmt"

type VpnInfo struct {
	Current int64 `json:"current" yaml:"current"`
	Limit   int64 `json:"limit" yaml:"limit"`
}

func NewVpmInfo() VpnInfo {
	return VpnInfo{
		Current: 0,
		Limit:   0,
	}
}

func (info VpnInfo) String() string {
	return fmt.Sprintf(
		"Current %d\nLimit: %d\n",
		info.Current,
		info.Limit)
}
