// Package rpc provides high level access to the Fantom Opera blockchain
// node through RPC interface.
package rpc

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	client "github.com/ethereum/go-ethereum/rpc"
	"log"
)

const (
	// headerObserverCapacity represents the capacity of new headers' observer channel
	headerObserverCapacity = 5000
)

// Rpc represents the implementation of the Blockchain interface for Fantom Opera node.
type Rpc struct {
	ftm *ethclient.Client
	// captured header queue
	headers chan *types.Header
}

// New creates a new instance of the RPC client.
func New(url string) *Rpc {
	c, err := connect(url)
	if err != nil {
		log.Fatalf("can not connect to the Opera node; %s", err.Error())
	}

	rpc := &Rpc{
		ftm:     ethclient.NewClient(c),
		headers: make(chan *types.Header, headerObserverCapacity),
	}

	return rpc
}

// connect opens RPC connection to the Opera node.
func connect(url string) (*client.Client, error) {
	c, err := client.Dial(url)
	if err != nil {
		return nil, err
	}
	return c, nil
}
