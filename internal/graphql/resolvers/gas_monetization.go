package resolvers

import (
	"ftm-gas-monetization/internal/repository"
	"github.com/ethereum/go-ethereum/common"
)

// GasMonetizationAddress returns the address of the gas monetization contract.
func (rs *RootResolver) GasMonetizationAddress() common.Address {
	return repository.R().GasMonetizationAddress()
}
