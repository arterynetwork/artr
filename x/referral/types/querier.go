package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

// Query endpoints supported by the referral querier
const (
	QueryStatus             = "status"
	QueryReferrer           = "referrer"
	QueryReferrals          = "referrals"
	QueryCoinsInNetwork     = "coins"
	QueryDelegatedInNetwork = "delegated"
	QueryCheckStatus        = "check-status"
)

type QueryResChildren []sdk.AccAddress

func (qr QueryResChildren) String() string {
	strs := make([]string, len(qr), len(qr))
	for i, adr := range qr {
		strs[i] = adr.String()
	}
	return strings.Join(strs[:], ", ")
}
