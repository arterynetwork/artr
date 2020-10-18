package types

import "fmt"

type ActivityInfo struct {
	Active   bool  `json:"active" yaml:"active"`
	ExpireAt int64 `json:"expire_at" yaml:"expire_at"`
}

func NewActivityInfo(active bool, expireAt int64) ActivityInfo {
	return ActivityInfo{
		Active:   active,
		ExpireAt: expireAt,
	}
}

func (info ActivityInfo) String() string {
	return fmt.Sprintf("Active: %t\nExpire at: %d", info.Active, info.ExpireAt)
}
