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
		common.HexToHash("0xa8f2a13a6c4c221e863c34b0174b2a8356551bc645dc295ae4b5796c240915aa"): handleProjectAdded,
		common.HexToHash("0x0c3ad6c6f2fc1e970caed51e87ac06c3a37569f33664f42771264a4ae8907822"): handleProjectSuspended,
		common.HexToHash("0x0737ed2cc6eb4cf4aefb6d1e1404305301a64cae58ccf508828a20412fb77f35"): handleProjectEnabled,
		common.HexToHash("0xf83ba82192ce64f0fd48145ca2b60956a005a2d4e28f14fb099fad71294b8ff3"): handleProjectContractAdded,
		common.HexToHash("0xd32f2e923c29ff9e7231f459d69add67f769d05c5069c23bbdea536fc0cf154a"): handleProjectContractRemoved,
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
			ProjectId: uint64(project.Id),
			Address:   &addr,
			Enabled:   true,
		}); err != nil {
			return fmt.Errorf("failed to add contract %s for project #%d: %v", contract.Hex(), project.ProjectId, err)
		}
	}
	return nil
}

// handleProjectSuspended is an event handler for the ProjectSuspended event.
func handleProjectSuspended(ctx context.Context, log *eth.Log, db *db.Db, repo *repository.Repository) error {
	if len(log.Data) != 32 || len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectSuspended", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectSuspended event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := db.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// suspend project
	activeTo := eventData["suspendedOnEpochNumber"].(*big.Int).Uint64()
	project.ActiveToEpoch = &activeTo
	err = db.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to suspend project #%d: %v", project.ProjectId, err)
	}
	return nil
}

// handleProjectEnabled is an event handler for the ProjectEnabled event.
func handleProjectEnabled(ctx context.Context, log *eth.Log, db *db.Db, repo *repository.Repository) error {
	if len(log.Data) != 32 || len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectEnabled", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectEnabled event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := db.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// enable project
	project.ActiveFromEpoch = eventData["enabledOnEpochNumber"].(*big.Int).Uint64()
	project.ActiveToEpoch = nil
	err = db.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to enable project #%d: %v", project.ProjectId, err)
	}
	return nil
}

// handleProjectContractAdded is an event handler for the ProjectContractAdded event.
func handleProjectContractAdded(ctx context.Context, log *eth.Log, db *db.Db, repo *repository.Repository) error {
	if len(log.Topics) != 3 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectContractAdded", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectContractAdded event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := db.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// add contract
	addr := types.Address{Address: common.BytesToAddress(log.Topics[2].Bytes())}
	if err := db.StoreProjectContract(ctx, &types.ProjectContract{
		ProjectId: uint64(project.Id),
		Address:   &addr,
		Enabled:   true,
	}); err != nil {
		return fmt.Errorf("failed to add contract %s for project #%d: %v", addr.Hex(), project.ProjectId, err)
	}

	return nil
}

// handleProjectContractRemoved is an event handler for the ProjectContractRemoved event.
func handleProjectContractRemoved(ctx context.Context, log *eth.Log, db *db.Db, repo *repository.Repository) error {
	if len(log.Topics) != 3 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectContractRemoved", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectContractRemoved event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// delete contract
	qb := db.ProjectContractQuery(ctx)
	addr := types.Address{Address: common.BytesToAddress(log.Topics[2].Bytes())}
	if err := qb.WhereAddress(&addr).Delete(); err != nil {
		return fmt.Errorf("failed to delete contract %s for project #%d: %v", addr.Hex(), log.Topics[1].Big().Uint64(), err)
	}
	return nil
}
