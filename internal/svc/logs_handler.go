// Package svc implements monitoring and scanning services of the API server.
package svc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	eth "github.com/ethereum/go-ethereum/core/types"
)

// handleProjectAdded is an event handler for the ProjectAdded event.
// It is called when a new project is added to the registry.
func handleProjectAdded(*eth.Log, *db.Db) {

}

func handleWithdrawalRequest(*eth.Log, *db.Db) {

}
