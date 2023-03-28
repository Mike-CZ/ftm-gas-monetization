package rpc

import (
	"math/big"
)

// CompleteWithdrawal completes withdrawal of the given amount from the given project.
func (rpc *Rpc) CompleteWithdrawal(projectId uint64, epoch uint64, amount *big.Int) error {
	_, err := rpc.dataProviderSession.CompleteWithdrawal(
		new(big.Int).SetUint64(projectId), new(big.Int).SetUint64(epoch), amount)
	return err
}
