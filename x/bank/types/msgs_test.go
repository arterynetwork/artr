// +build testing

package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgSendRoute(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("from"))
	addr2 := sdk.AccAddress([]byte("to"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	var msg = NewMsgSend(addr1, addr2, coins)

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "send")
}

func TestMsgSendValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("from0000000000000000"))
	addr2 := sdk.AccAddress([]byte("to000000000000000000"))
	atom123 := sdk.NewCoins(sdk.NewInt64Coin("atom", 123))
	atom0 := sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
	atom123eth123 := sdk.NewCoins(sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 123))
	atom123eth0 := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 0)}

	var emptyAddr sdk.AccAddress

	cases := []struct {
		valid bool
		tx    MsgSend
	}{
		{true, *NewMsgSend(addr1, addr2, atom123)},       // valid send
		{true, *NewMsgSend(addr1, addr2, atom123eth123)}, // valid send with multiple coins
		{false, *NewMsgSend(addr1, addr2, atom0)},        // non positive coin
		{false, *NewMsgSend(addr1, addr2, atom123eth0)},  // non positive coin in multicoins
		{false, *NewMsgSend(emptyAddr, addr2, atom123)},  // empty from addr
		{false, *NewMsgSend(addr1, emptyAddr, atom123)},  // empty to addr
	}

	for i, tc := range cases {
		err := tc.tx.ValidateBasic()
		if tc.valid {
			require.NoError(t, err, "Case# %d: %+v", i, tc.tx)
		} else {
			require.Error(t, err, "Case #%d: %+v", i, tc.tx)
		}
	}
}

func TestMsgSendGetSignBytes(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("input"))
	addr2 := sdk.AccAddress([]byte("output"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	var msg = NewMsgSend(addr1, addr2, coins)
	res := msg.GetSignBytes()

	//expected := `{"type":"cosmos-sdk/MsgSend","value":{"amount":[{"amount":"10","denom":"atom"}],"from_address":"cosmos1d9h8qat57ljhcm","to_address":"cosmos1da6hgur4wsmpnjyg"}}`
	expected := []byte("\x0A\x15cosmos1d9h8qat57ljhcm\x12\x17cosmos1da6hgur4wsmpnjyg\x1A\x0A\x0A\x04atom\x12\x0210")
	require.Equal(t, expected, res)
}

func TestMsgSendGetSigners(t *testing.T) {
	var msg = NewMsgSend(sdk.AccAddress([]byte("input1______________")), sdk.AccAddress{}, sdk.NewCoins())
	res := msg.GetSigners()
	// TODO: fix this !
	require.Equal(t, fmt.Sprintf("%v", res), "[696E707574315F5F5F5F5F5F5F5F5F5F5F5F5F5F]")
}

func TestInputValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte{1, 2})
	addr2 := sdk.AccAddress([]byte{7, 8})
	someCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123))
	multiCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 20))

	var emptyAddr sdk.AccAddress
	emptyCoins := sdk.NewCoins()
	emptyCoins2 := sdk.NewCoins(sdk.NewInt64Coin("eth", 0))
	someEmptyCoins := sdk.Coins{sdk.NewInt64Coin("eth", 10), sdk.NewInt64Coin("atom", 0)}
	unsortedCoins := sdk.Coins{sdk.NewInt64Coin("eth", 1), sdk.NewInt64Coin("atom", 1)}

	cases := []struct {
		valid bool
		txIn  Input
	}{
		// auth works with different apps
		{true, NewInput(addr1, someCoins)},
		{true, NewInput(addr2, someCoins)},
		{true, NewInput(addr2, multiCoins)},

		{false, NewInput(emptyAddr, someCoins)},  // empty address
		{false, NewInput(addr1, emptyCoins)},     // invalid coins
		{false, NewInput(addr1, emptyCoins2)},    // invalid coins
		{false, NewInput(addr1, someEmptyCoins)}, // invalid coins
		{false, NewInput(addr1, unsortedCoins)},  // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txIn.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}

func TestOutputValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte{1, 2})
	addr2 := sdk.AccAddress([]byte{7, 8})
	someCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123))
	multiCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 20))

	var emptyAddr sdk.AccAddress
	emptyCoins := sdk.NewCoins()
	emptyCoins2 := sdk.NewCoins(sdk.NewInt64Coin("eth", 0))
	someEmptyCoins := sdk.Coins{sdk.NewInt64Coin("eth", 10), sdk.NewInt64Coin("atom", 0)}
	unsortedCoins := sdk.Coins{sdk.NewInt64Coin("eth", 1), sdk.NewInt64Coin("atom", 1)}

	cases := []struct {
		valid bool
		txOut Output
	}{
		// auth works with different apps
		{true, NewOutput(addr1, someCoins)},
		{true, NewOutput(addr2, someCoins)},
		{true, NewOutput(addr2, multiCoins)},

		{false, NewOutput(emptyAddr, someCoins)},  // empty address
		{false, NewOutput(addr1, emptyCoins)},     // invalid coins
		{false, NewOutput(addr1, emptyCoins2)},    // invalid coins
		{false, NewOutput(addr1, someEmptyCoins)}, // invalid coins
		{false, NewOutput(addr1, unsortedCoins)},  // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txOut.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}
