// Package svc implements monitoring and scanning services of the API server.
package svc

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	eth "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// EventHandler represents a function used to process event log record.
type EventHandler func(context.Context, *eth.Log, *db.Db, *repository.Repository) error

// topics represents a map of topics to their respective event handlers.
func topics() map[common.Hash]EventHandler {
	return map[common.Hash]EventHandler{
		/* ProjectAdded(uint256, address, address, string, uint256, address[])*/
		common.HexToHash("0xa8f2a13a6c4c221e863c34b0174b2a8356551bc645dc295ae4b5796c240915aa"): handleProjectAdded,
	}
}

// handleProjectAdded is an event handler for the ProjectAdded event.
// It is called when a new project is added to the registry.
func handleProjectAdded(ctx context.Context, log *eth.Log, db *db.Db, repo *repository.Repository) error {
	if len(log.Data) < 192 || len(log.Topics) != 4 {
		return fmt.Errorf("not ProjectAdded() event #%d/#%d; expected 192 bytes of data, %d given; expected 4 topics, %d given",
			log.BlockNumber, log.Index, len(log.Data), len(log.Topics))
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectAdded", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectAdded event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// create project
	ownerAddr := types.Address{Address: common.BytesToAddress(log.Topics[2].Bytes())}
	receiverAddr := types.Address{Address: common.BytesToAddress(log.Topics[3].Bytes())}
	project := &types.Project{
		ProjectId:           log.Topics[1].Big().Uint64(),
		OwnerAddress:        &ownerAddr,
		ReceiverAddress:     &receiverAddr,
		LastWithdrawalEpoch: nil,
		ActiveFromEpoch:     eventData["activeFromEpoch"].(*big.Int).Uint64(),
		ActiveToEpoch:       nil,
	}
	// store project
	if err := db.StoreProject(ctx, project); err != nil {
		return fmt.Errorf("failed to add project #%d: %v", project.ProjectId, err)
	}
	// create contracts
	for _, contract := range eventData["contracts"].([]common.Address) {
		addr := types.Address{Address: contract}
		if err := db.StoreProjectContract(ctx, &types.ProjectContract{
			ProjectId: project.ProjectId,
			Address:   &addr,
			Enabled:   true,
		}); err != nil {
			return fmt.Errorf("failed to add contract %s for project #%d: %v", contract.Hex(), project.ProjectId, err)
		}
	}

	return nil
}
