package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/app"
)

func main() {
	cmd := &cobra.Command{
		Use:   "artrtxdec",
		Short: "Small handy tool for Artery Network transactions decoding",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use: "tx <tx>",
			Short: "Parse transaction base64",
			Args: cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error { return tx(args[0]) },
		},
		&cobra.Command{
			Use: "acc <acc>",
			Short: "Convert an account address from hex to bech32",
			Args: cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error { return acc(args[0]) },
		},
	)

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func tx(base64Str string) error {
	cdc := app.MakeCodec()
	decoder := authTypes.DefaultTxDecoder(cdc)

	bz, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return errors.Wrap(err, "cannot decode base64")
	}

	tx, err := decoder(bz)
	if err != nil {
		return errors.Wrap(err, "cannot decode tx")
	}
	msgs := tx.GetMsgs()
	for i, msg := range msgs {
		fmt.Printf("#%d Msg Type: %s\n%+v\n", i, msg.Type(), msg)
	}
	return nil
}

func acc(hexStr string) error {
	app.InitConfig()

	var (
		addr sdk.AccAddress
		err error
	)
	addr, err = hex.DecodeString(hexStr)
	if err != nil {
		return errors.Wrap(err, "cannot decode hex string")
	}
	fmt.Println(addr.String())
	return nil
}