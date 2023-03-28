package repository

import "math/big"

// CompleteWithdrawal completes withdrawal of the given amount from the given project.
func (repo *Repository) CompleteWithdrawal(projectId uint64, epoch uint64, amount *big.Int) error {
	return repo.rpc.CompleteWithdrawal(projectId, epoch, amount)
}
