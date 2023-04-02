package repository

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// CompleteWithdrawal completes withdrawal of the given amount from the given project.
func (repo *Repository) CompleteWithdrawal(projectId uint64, epoch uint64, amount *big.Int) error {
	return repo.rpc.CompleteWithdrawal(projectId, epoch, amount)
}

// HasPendingWithdrawal returns true if there is a pending withdrawal for the given project.
func (repo *Repository) HasPendingWithdrawal(projectId uint64, epoch uint64) (bool, error) {
	return repo.rpc.HasPendingWithdrawal(projectId, epoch)
}

// GasMonetizationAddress returns the address of the gas monetization contract.
func (repo *Repository) GasMonetizationAddress() common.Address {
	return repo.rpc.GasMonetizationAddress()
}
