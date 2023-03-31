package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// CompleteWithdrawal completes withdrawal of the given amount from the given project.
func (rpc *Rpc) CompleteWithdrawal(projectId uint64, epoch uint64, amount *big.Int) error {
	_, err := rpc.dataProviderSession.CompleteWithdrawal(
		new(big.Int).SetUint64(projectId), new(big.Int).SetUint64(epoch), amount)
	return err
}

// GasMonetizationAddress returns the address of the gas monetization contract.
func (rpc *Rpc) GasMonetizationAddress() common.Address {
	return rpc.gasMonetizationAddress
}

// SetGasMonetizationAddress sets the address of the gas monetization contract.
// This is used for testing purposes only.
func (rpc *Rpc) SetGasMonetizationAddress(addr common.Address) {
	rpc.gasMonetizationAddress = addr
}
