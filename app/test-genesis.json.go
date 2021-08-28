// +build testing

package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var DefaultGenesisUsers map[string]sdk.AccAddress
var NonExistingUser sdk.AccAddress

func initDefaultGenesisUsers() {
	DefaultGenesisUsers = map[string]sdk.AccAddress{
		//                 1
		//         ┌───────┴───────┐
		//         2               3
		//     ┌───┴───┐       ┌───┴───┐
		//     4       5       6       7
		//  ┌──┴──┐ ┌──┴──┐ ┌──┴──┐ ┌──┴──┐
		//  8     9 A     B C     D E     F
		"user1":  accAddr("artrt1d4ezqdj03uachct8hum0z9zlfftzdq2f7x0fwf"),
		"user2":  accAddr("artrt1h8s8yf433ypjc5htavsyc9zvg3vk43vm5du8gy"),
		"user3":  accAddr("artrt1cjqvu8pns5ff3vcy4r7qwy57f2ts8chsk22na8"),
		"user4":  accAddr("artrt1hdayszxl2ahw4rm0mct72rxzukq058mac6zgjp"),
		"user5":  accAddr("artrt15n7wt45x4tkgunp25wylrjymjnkqug80eu2h2g"),
		"user6":  accAddr("artrt1sqh7ly9z3yme0k32f42qu330a663zkcumetmem"),
		"user7":  accAddr("artrt1uaz24ndash8umld4xpfn3zpknk5ske3xghy9yq"),
		"user8":  accAddr("artrt14qn52aqd5dp4eycngm9e5ryrqdaf0u2ze6vqg7"),
		"user9":  accAddr("artrt1h2emj28qqj0e4k3azyzqdqznxdkf9r5529dtr5"),
		"user10": accAddr("artrt1fedl94g9gqnntzqtgmxyp6msztzvw435hw8v69"),
		"user11": accAddr("artrt1j9j50h6k3v2p70nar6234etd2amxdeysra0dun"),
		"user12": accAddr("artrt1tkl5vyca6mlfhl0zmjkl8nkcmlulpre5ravkhl"),
		"user13": accAddr("artrt1xkwt5k2pktltp0jzk6hjz9k2k89l3044rfns6m"),
		"user14": accAddr("artrt1q2w5ytm97g490lcux69n3vfprqsdv65v0r2k9g"),
		"user15": accAddr("artrt1j29a9493fmlkjr9hmp54ltjun2meph9lst6c3j"),
		"root":   accAddr("artrt1yhy6d3m4utltdml7w7zte7mqx5wyuskqppw34n"),
	}
	NonExistingUser = accAddr("artrt1t8h48rk0wyvuvdae5aysmlnfly6rpqe6r77rdd")
}
func accAddr(s string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return addr
}

const DefaultUser1ConsPubKey = "artrtvalconspub1zcjduepqpme87trszw7awc62ra2de9edwr40v7xy7yfhvpvds96fncagm04qv3rnr2"
