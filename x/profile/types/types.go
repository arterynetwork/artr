package types

import (
	"fmt"
	"strings"
)

type Profile struct {
	AutoPay        bool   `json:"autopay" yaml:"autopay"`
	ActiveUntil    uint64 `json:"active_until" yaml:"active_until"`
	Noding         bool   `json:"noding" yaml:"noding"`
	Storage        bool   `json:"storage" yaml:"storage"`
	Validator      bool   `json:"validator" yaml:"validator"`
	VPN            bool   `json:"VPN" yaml:"VPN"`
	Nickname       string `json:"nickname" yaml:"nickname"`
	CardNumber     uint64 `json:"card_number,omitempty" yaml:"card_number"`
}

func (p Profile) String() string {
	return strings.TrimSpace(fmt.Sprintf(
			"AutoPayment: %t\n"+
			"ActiveUntil: %d\n"+
			"Noding: %t\n"+
			"Storage: %t\n"+
			"VPN: %t\n"+
			"Validator: %t\n"+
			"Nilname: %s\n"+
			"CardNumber: %012d",
		p.AutoPay,
		p.ActiveUntil,
		p.Noding,
		p.Storage,
		p.VPN,
		p.Validator,
		p.Nickname,
		p.CardNumber))
}
