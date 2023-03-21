// Package rpc provides high level access to the Fantom Opera blockchain
// node through RPC interface.
package rpc

import "github.com/ethereum/go-ethereum/accounts/abi"

// GasMonetizationAbi provides access to decoded ABI of Fantom Gas Monetization contract.
func (rpc *Rpc) GasMonetizationAbi() *abi.ABI {
	return rpc.abiGasMonetization
}
