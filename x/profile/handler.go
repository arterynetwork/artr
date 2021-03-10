package profile

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/arterynetwork/artr/x/profile/types"
)

// NewHandler creates an sdk.Handler for all the profile type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgSetProfile:
			return handleMsgSetProfile(ctx, k, msg)
		case types.MsgCreateAccount:
			return handleMsgCreateAccount(ctx, k, msg)
		case types.MsgCreateAccountWithProfile:
			return handleMsgCreateAccountWithProfile(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// Handle a message to set profile
func handleMsgSetProfile(ctx sdk.Context, keeper Keeper, msg types.MsgSetProfile) (*sdk.Result, error) {
	// validate nickname, if not empty
	if strings.TrimSpace(msg.Profile.Nickname) != "" {
		if len(msg.Profile.Nickname) < 3 {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "nick too short")
		}

		if strings.ContainsAny(msg.Profile.Nickname, " */:'\"=[],.") {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "nickname contains invalid characters")
		}
	}

	if err := keeper.SetProfile(ctx, msg.Address, msg.Profile); err != nil {
		return new(sdk.Result), errors.Wrap(err, "cannot set profile")
	}
	return &sdk.Result{}, nil
}

// Handle a message to set profile
func handleMsgCreateAccount(ctx sdk.Context, keeper Keeper, msg types.MsgCreateAccount) (*sdk.Result, error) {
	// check if account exists
	account := keeper.AccountKeeper.GetAccount(ctx, msg.NewAccount)

	if account != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s is already exist", msg.NewAccount)
	}

	creator := keeper.AccountKeeper.GetAccount(ctx, msg.Address)

	if creator == nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress, "Creator account %s is invalid", msg.Address,
		)
	}

	p := keeper.GetParams(ctx)

	freeCreation := false

	for _, acc := range p.Creators {
		if acc.Equals(msg.Address) {
			freeCreation = true
			break
		}
	}

	if !freeCreation {
		oldCoins := creator.GetCoins()

		amt := sdk.NewCoins(sdk.NewCoin(types.MainDenom, sdk.NewInt(p.Fee)))

		_, hasNeg := oldCoins.SafeSub(amt)

		if hasNeg {
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds, "insufficient account funds; %s < %s", oldCoins, amt,
			)
		}

		err := keeper.SupplyKeeper.SendCoinsFromAccountToModule(ctx, msg.Address, auth.FeeCollectorName, amt)

		if err != nil {
			return nil, err
		}
	}

	if err := keeper.CreateAccount(ctx, msg.NewAccount, msg.ReferralAddress); err != nil {
		return new(sdk.Result), errors.Wrap(err, "cannot create account")
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// Handle a message to set profile
func handleMsgCreateAccountWithProfile(ctx sdk.Context, keeper Keeper, msg types.MsgCreateAccountWithProfile) (*sdk.Result, error) {
	// check if account exists
	account := keeper.AccountKeeper.GetAccount(ctx, msg.NewAccount)

	if account != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s is already exist", msg.NewAccount)
	}

	creator := keeper.AccountKeeper.GetAccount(ctx, msg.Address)

	if creator == nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress, "Creator account %s is invalid", msg.Address,
		)
	}

	p := keeper.GetParams(ctx)

	freeCreation := false

	for _, acc := range p.Creators {
		if acc.Equals(msg.Address) {
			freeCreation = true
			break
		}
	}

	if !freeCreation {
		oldCoins := creator.GetCoins()

		amt := sdk.NewCoins(sdk.NewCoin(types.MainDenom, sdk.NewInt(p.Fee)))

		_, hasNeg := oldCoins.SafeSub(amt)

		if hasNeg {
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds, "insufficient account funds; %s < %s", oldCoins, amt,
			)
		}

		err := keeper.SupplyKeeper.SendCoinsFromAccountToModule(ctx, msg.Address, auth.FeeCollectorName, amt)

		if err != nil {
			return nil, err
		}
	}

	if err := keeper.CreateAccountWithProfile(ctx, msg.NewAccount, msg.ReferralAddress, msg.Profile); err != nil {
		return new(sdk.Result), errors.Wrap(err, "cannot create account")
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
