package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Input models transaction input
type Input struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Coins   sdk.Coins      `json:"coins" yaml:"coins"`
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() error {
	if len(in.Address) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "input address missing")
	}
	if !in.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}
	if !in.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}
	return nil
}

// NewInput - create a transaction input, used with MsgMultiSend
func NewInput(addr sdk.AccAddress, coins sdk.Coins) Input {
	return Input{
		Address: addr,
		Coins:   coins,
	}
}

// Output models transaction outputs
type Output struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Coins   sdk.Coins      `json:"coins" yaml:"coins"`
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() error {
	if len(out.Address) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "output address missing")
	}
	if !out.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}
	if !out.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}
	return nil
}

// NewOutput - create a transaction output, used with MsgMultiSend
func NewOutput(addr sdk.AccAddress, coins sdk.Coins) Output {
	return Output{
		Address: addr,
		Coins:   coins,
	}
}

// ValidateInputsOutputs validates that each respective input and output is
// valid and that the sum of inputs is equal to the sum of outputs.
func ValidateInputsOutputs(inputs []Input, outputs []Output) error {
	var totalIn, totalOut sdk.Coins

	for _, in := range inputs {
		if err := in.ValidateBasic(); err != nil {
			return err
		}

		totalIn = totalIn.Add(in.Coins...)
	}

	for _, out := range outputs {
		if err := out.ValidateBasic(); err != nil {
			return err
		}

		totalOut = totalOut.Add(out.Coins...)
	}

	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInputOutputMismatch
	}

	return nil
}
