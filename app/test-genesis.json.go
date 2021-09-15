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
		"user1":  accAddr("artr1d4ezqdj03uachct8hum0z9zlfftzdq2f6yzvhj"),
		"user2":  accAddr("artr1h8s8yf433ypjc5htavsyc9zvg3vk43vms03z3l"),
		"user3":  accAddr("artr1cjqvu8pns5ff3vcy4r7qwy57f2ts8chsjg8kyu"),
		"user4":  accAddr("artr1hdayszxl2ahw4rm0mct72rxzukq058mauc0dt6"),
		"user5":  accAddr("artr15n7wt45x4tkgunp25wylrjymjnkqug80a78jnn"),
		"user6":  accAddr("artr1sqh7ly9z3yme0k32f42qu330a663zkculmx7qq"),
		"user7":  accAddr("artr1uaz24ndash8umld4xpfn3zpknk5ske3xv4fqam"),
		"user8":  accAddr("artr14qn52aqd5dp4eycngm9e5ryrqdaf0u2zacp939"),
		"user9":  accAddr("artr1h2emj28qqj0e4k3azyzqdqznxdkf9r55w8qw60"),
		"user10": accAddr("artr1fedl94g9gqnntzqtgmxyp6msztzvw435nv2fr7"),
		"user11": accAddr("artr1j9j50h6k3v2p70nar6234etd2amxdeys8lzg9g"),
		"user12": accAddr("artr1tkl5vyca6mlfhl0zmjkl8nkcmlulpre58lpnwy"),
		"user13": accAddr("artr1xkwt5k2pktltp0jzk6hjz9k2k89l30448t74rq"),
		"user14": accAddr("artr1q2w5ytm97g490lcux69n3vfprqsdv65vtp8nun"),
		"user15": accAddr("artr1j29a9493fmlkjr9hmp54ltjun2meph9l5fhagf"),
		"root":   accAddr("artr1yhy6d3m4utltdml7w7zte7mqx5wyuskq9rr5vg"),
	}
	NonExistingUser = accAddr("artr1t8h48rk0wyvuvdae5aysmlnfly6rpqe68unx5k")
}
func accAddr(s string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return addr
}

const DefaultUser1ConsPubKey = "artrvalconspub1zcjduepqpme87trszw7awc62ra2de9edwr40v7xy7yfhvpvds96fncagm04qxu308e"
