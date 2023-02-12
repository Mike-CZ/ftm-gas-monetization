// Package rpc provides high level access to the Fantom Opera blockchain
// node through RPC interface.
package rpc

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	client "github.com/ethereum/go-ethereum/rpc"
	"log"
)

// Rpc represents the implementation of the Blockchain interface for Fantom Opera node.
type Rpc struct {
	ftm *ethclient.Client

	// captured header queue
	headers chan *types.Header
}

func New() *Rpc {
	client, err := connect("http://localhost:8545")
	if err != nil {
		log.Fatalf("can not connect to the Opera node; %s", err.Error())
	}

}

// connects opens RPC connection to the Opera node.
func connect(url string) (*client.Client, error) {
	c, err := client.Dial(url)
	if err != nil {
		return nil, err
	}
	return c, nil
}
