package types

import (
	"github.com/arterynetwork/artr/x/referral"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

const (
	ProposalTypeNone = 0
	//1. Стоимость входа
	ProposalTypeEnterPrice = 1
	//2. Вознаграждение за делегирование
	ProposalTypeDelegationAward = 2
	//3. Сетевое вознаграждение с делегирования
	ProposalTypeDelegationNetworkAward = 3
	//4. Сетевое вознаграждение с покупки продукта
	ProposalTypeProductNetworkAward = 4
	//5. Состав голосующего совета (включение / исключение участника)
	ProposalTypeGovernmentAdd    = 5
	ProposalTypeGovernmentRemove = 6
	//6. Стоимость VPN / гб сверх тарифа
	ProposalTypeProductVpnBasePrice = 7
	//7. Стоимость storage / гб сверх тарифа
	ProposalTypeProductStorageBasePrice = 9
	// Аккаунт с правом бесплатного создания новых аккаунтов
	ProposalTypeAddFreeCreator    = 10
	ProposalTypeRemoveFreeCreator = 11
	// Обновление blockchain
	ProposalTypeSoftwareUpgrade       = 12
	ProposalTypeCancelSoftwareUpgrade = 13
	// Служебные (с иммунитетом к изменению статуса и делегации) валидаторы
	ProposalTypeStaffValidatorAdd    = 14
	ProposalTypeStaffValidatorRemove = 15
	// Кто может подписывать транзакции на выплату вознаграждений за хранилище/VPN
	ProposalTypeEarningSignerAdd    = 16
	ProposalTypeEarningSignerRemove = 17
	// Кто может менять курс монетки
	ProposalTypeRateChangeSignerAdd    = 18
	ProposalTypeRateChangeSignerRemove = 19
	// Кто считает трафик VPN и актуализирует данные по нему
	ProposalTypeVpnCurrentSignerAdd    = 20
	ProposalTypeVpnCurrentSignerRemove = 21
	// Стоимость переноса аккаунта к другому пригласившему
	ProposalTypeTransitionCost = 22
	// Минимальная сумма перевода
	ProposalTypeMinSend = 23
	// Минимальная сумма делегирования
	ProposalTypeMinDelegate = 24
	// Максимальное количество валидаторов
	ProposalTypeMaxValidators = 25
	// Амнистия
	ProposalTypeGeneralAmnesty = 26
	// "Счастливые" валидаторы
	ProposalTypeLotteryValidators = 27
	// Статус, начиная с которого доступна валидация
	ProposalTypeValidatorMinimalStatus = 28
)

// EmptyProposalParams

var _ ProposalParams = &EmptyProposalParams{}

type EmptyProposalParams struct {
}

func (e EmptyProposalParams) String() string {
	return "EmptyProposalParams"
}

// PriceProposalParams

var _ ProposalParams = &PriceProposalParams{}

type PriceProposalParams struct {
	Price uint32 `json:"price" yaml:"price"`
}

func (p PriceProposalParams) String() string {
	return fmt.Sprintf("Price: %d", p.Price)
}

// DelegationAwardProposalParams

var _ ProposalParams = &DelegationAwardProposalParams{}

type DelegationAwardProposalParams struct {
	Minimal      uint8 `json:"minimal" yaml:"minimal"`
	ThousandPlus uint8 `json:"thousand_plus" yaml:"thousand_plus"`
	TenKPlus     uint8 `json:"ten_k_plus" yaml:"ten_k_plus"`
	HundredKPlus uint8 `json:"hundred_k_plus" yaml:"hundred_k_plus"`
}

func (p DelegationAwardProposalParams) String() string {
	return fmt.Sprintf("Minimal: %d; 1K+: %d; 10K+: %d; 100K+: %d", p.Minimal, p.ThousandPlus, p.TenKPlus, p.ThousandPlus)
}

// NetworkAwardProposalParams
var _ ProposalParams = &NetworkAwardProposalParams{}

type NetworkAwardProposalParams struct {
	Award referral.NetworkAward `json:"award" yaml:"award"`
}

func (p NetworkAwardProposalParams) String() string {
	builder := strings.Builder{}
	builder.WriteString("Company: ")
	builder.WriteString(p.Award.Company.String())
	for i := 0; i < 10; i++ {
		builder.WriteString(fmt.Sprintf("; Level %d: %s", i+1, p.Award.Network[i]))
	}
	return builder.String()
}

// AddressProposalParams

var _ ProposalParams = &AddressProposalParams{}

type AddressProposalParams struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

func (params AddressProposalParams) String() string {
	return "Address: " + params.Address.String()
}

// SoftwareUpgradeProposalParams

var _ ProposalParams = &SoftwareUpgradeProposalParams{}

type SoftwareUpgradeProposalParams struct {
	// Name - upgrade name
	Name string `json:"name" yaml:"name"`
	// Height - block height to schedule the upgrade at
	Height int64 `json:"height" yaml:"height"`
	// Binaries - a link (with a checksum) to a JSON file containing upgrade data (binary URIs and so on)
	// Please refer to https://github.com/regen-network/cosmosd#auto-download
	Info string `json:"binaries" yaml:"binaries"`
}

func (p SoftwareUpgradeProposalParams) String() string {
	return fmt.Sprintf("Name: %s; Height: %d; Binaries: %s", p.Name, p.Height, p.Info)
}

// MinAmountProposalParams

var _ ProposalParams = &MinAmountProposalParams{}

type MinAmountProposalParams struct {
	MinAmount int64 `json:"min_amount" yaml:"min_amount"`
}

func (params MinAmountProposalParams) String() string {
	return fmt.Sprintf("MinAmount: %d", params.MinAmount)
}

// ShortCountProposalParams

var _ ProposalParams = &ShortCountProposalParams{}

type ShortCountProposalParams struct {
	Count uint16 `json:"count" yaml:"count"`
}

func (params ShortCountProposalParams) String() string {
	return fmt.Sprintf("Count: %d", params.Count)
}

var _ ProposalParams = &StatusProposalParams{}

type StatusProposalParams struct {
	Status uint8 `json:"status" yaml:"status"`
}

func (p StatusProposalParams) String() string {
	return fmt.Sprintf("Status: %d", p.Status)
}
