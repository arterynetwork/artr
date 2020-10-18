package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProposalParams interface {
	String() string
}

type Proposal struct {
	Name     string         `json:"name" yaml:"name"`
	TypeCode uint8          `json:"type_code" yaml:"type_code"`
	Params   ProposalParams `json:"params" yaml:"params"`
	Author   sdk.AccAddress `json:"author" yaml:"author"`
	EndBlock int64          `json:"end_block" yaml:"end_block"`
}

func (p Proposal) String() string {
	return fmt.Sprintf("Name: %s\nTypeCode: %d\nParams: %s\n",
		p.Name, p.TypeCode, p.Params.String())
}

type Government []sdk.AccAddress

func NewEmptyGovernment() Government {
	return make(Government, 0)
}

func NewGovernment(accounts []sdk.AccAddress) Government {
	return Government(accounts)
}

func (g Government) String() string {
	return fmt.Sprint([]sdk.AccAddress(g))
}

func (g Government) Contains(addr sdk.AccAddress) bool {
	for _, elem := range g {
		if addr.Equals(elem) {
			return true
		}
	}

	return false
}

func (g Government) Remove(addr sdk.AccAddress) Government {
	for index, elem := range g {
		if addr.Equals(elem) {
			g = append(g[:index], g[index+1:]...)
			return g
		}
	}

	return g
}

func (g Government) Append(addr sdk.AccAddress) Government {
	g = append(g, addr)
	return g
}

type ProposalHistoryRecord struct {
	Proposal   Proposal   `json:"proposal" yaml:"proposal"`
	Government Government `json:"government" yaml:"government"`
	Agreed     Government `json:"agreed" yaml:"agreed"`
	Disagreed  Government `json:"disagreed" yaml:"disagreed"`
	Started    int64      `json:"started" yaml:"started"`
	Ended      int64      `json:"ended" yaml:"ended"`
}
