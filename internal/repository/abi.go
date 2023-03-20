package repository

import "github.com/ethereum/go-ethereum/accounts/abi"

// GasMonetizationAbi provides access to decoded ABI of Fantom Gas Monetization contract.
func (repo *Repository) GasMonetizationAbi() *abi.ABI {
	return repo.rpc.GasMonetizationAbi()
}
