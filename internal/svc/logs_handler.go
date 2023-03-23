// Package svc implements monitoring and scanning services of the API server.
package svc

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	eth "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// EventHandler represents a function used to process event log record.
type EventHandler func(context.Context, *eth.Log, *db.Db) error

// initializeTopics represents a map of topics to their respective event handlers.
func (bld *blkDispatcher) initializeTopics() {
	bld.topics = map[common.Hash]EventHandler{
		common.HexToHash("0xa8f2a13a6c4c221e863c34b0174b2a8356551bc645dc295ae4b5796c240915aa"): bld.handleProjectAdded,
		common.HexToHash("0x0c3ad6c6f2fc1e970caed51e87ac06c3a37569f33664f42771264a4ae8907822"): bld.handleProjectSuspended,
		common.HexToHash("0x0737ed2cc6eb4cf4aefb6d1e1404305301a64cae58ccf508828a20412fb77f35"): bld.handleProjectEnabled,
		common.HexToHash("0xf83ba82192ce64f0fd48145ca2b60956a005a2d4e28f14fb099fad71294b8ff3"): bld.handleProjectContractAdded,
		common.HexToHash("0xd32f2e923c29ff9e7231f459d69add67f769d05c5069c23bbdea536fc0cf154a"): bld.handleProjectContractRemoved,
		common.HexToHash("0x781779743e625d6e652139cabc7e7c736ad376a0f1302b1b5c346548d948c72e"): bld.handleProjectMetadataUriUpdated,
		common.HexToHash("0xc96c5102d284d786d29b5d0d7dda6ce493724355b762993adfef62b7220f161c"): bld.handleProjectRecipientUpdated,
		common.HexToHash("0xffc579e983741c17a95792c458e2ae8c933b1bf7f5cd84f3bca571505c25d42a"): bld.handleProjectOwnerUpdated,
		common.HexToHash("0x6b19bb08027e5bee64cbe3f99bbbfb671c0e134643993f0ad046fd01d020b342"): bld.handleWithdrawalRequest,
	}
}

// handleProjectAdded is an event handler for the ProjectAdded event.
// It is called when a new project is added to the registry.
func (bld *blkDispatcher) handleProjectAdded(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Data) < 192 || len(log.Topics) != 4 {
		return fmt.Errorf("not ProjectAdded() event #%d/#%d; expected 192 bytes of data, %d given; expected 4 topics, %d given",
			log.BlockNumber, log.Index, len(log.Data), len(log.Topics))
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectAdded", log.Data)
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
	if err := transaction.StoreProject(ctx, project); err != nil {
		return fmt.Errorf("failed to add project #%d: %v", project.ProjectId, err)
	}
	// create contracts
	for _, contract := range eventData["contracts"].([]common.Address) {
		addr := types.Address{Address: contract}
		if err := transaction.StoreProjectContract(ctx, &types.ProjectContract{
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
func (bld *blkDispatcher) handleProjectSuspended(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Data) != 32 || len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectSuspended", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectSuspended event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := transaction.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// suspend project
	activeTo := eventData["suspendedOnEpochNumber"].(*big.Int).Uint64()
	project.ActiveToEpoch = &activeTo
	err = transaction.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to suspend project #%d: %v", project.ProjectId, err)
	}
	return nil
}

// handleProjectEnabled is an event handler for the ProjectEnabled event.
func (bld *blkDispatcher) handleProjectEnabled(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Data) != 32 || len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectEnabled", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectEnabled event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := transaction.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// enable project
	project.ActiveFromEpoch = eventData["enabledOnEpochNumber"].(*big.Int).Uint64()
	project.ActiveToEpoch = nil
	err = transaction.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to enable project #%d: %v", project.ProjectId, err)
	}
	return nil
}

// handleProjectContractAdded is an event handler for the ProjectContractAdded event.
func (bld *blkDispatcher) handleProjectContractAdded(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Topics) != 3 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectContractAdded", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectContractAdded event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := transaction.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// add contract
	addr := types.Address{Address: common.BytesToAddress(log.Topics[2].Bytes())}
	if err := transaction.StoreProjectContract(ctx, &types.ProjectContract{
		ProjectId: uint64(project.Id),
		Address:   &addr,
		Enabled:   true,
	}); err != nil {
		return fmt.Errorf("failed to add contract %s for project #%d: %v", addr.Hex(), project.ProjectId, err)
	}

	return nil
}

// handleProjectContractRemoved is an event handler for the ProjectContractRemoved event.
func (bld *blkDispatcher) handleProjectContractRemoved(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Topics) != 3 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectContractRemoved", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectContractRemoved event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// delete contract
	qb := transaction.ProjectContractQuery(ctx)
	addr := types.Address{Address: common.BytesToAddress(log.Topics[2].Bytes())}
	if err := qb.WhereAddress(&addr).Delete(); err != nil {
		return fmt.Errorf("failed to delete contract %s for project #%d: %v", addr.Hex(), log.Topics[1].Big().Uint64(), err)
	}
	return nil
}

// handleProjectMetadataUriUpdated is an event handler for the ProjectMetadataUriUpdated event.
func (bld *blkDispatcher) handleProjectMetadataUriUpdated(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectMetadataUriUpdated", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectMetadataUriUpdated event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// TODO: update metadata
	//projectId := log.Topics[1].Big().Uint64()
	//uri := eventData["metadataUri"].(string)
	return nil
}

// handleProjectRecipientUpdated is an event handler for the ProjectRewardsRecipientUpdated event.
func (bld *blkDispatcher) handleProjectRecipientUpdated(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectRewardsRecipientUpdated", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectRewardsRecipientUpdated event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := transaction.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// update recipient
	recipient := types.Address{Address: eventData["recipient"].(common.Address)}
	project.ReceiverAddress = &recipient
	err = transaction.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to update recipient %s for project #%d: %v", recipient.Hex(), project.ProjectId, err)
	}
	return nil
}

// handleProjectOwnerUpdated is an event handler for the ProjectOwnerUpdated event.
func (bld *blkDispatcher) handleProjectOwnerUpdated(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "ProjectOwnerUpdated", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectOwnerUpdated event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := transaction.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// update owner
	owner := types.Address{Address: eventData["owner"].(common.Address)}
	project.OwnerAddress = &owner
	err = transaction.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to update owner %s for project #%d: %v", owner.Hex(), project.ProjectId, err)
	}
	return nil
}

// handleWithdrawalRequest is an event handler for the WithdrawalRequested event.
func (bld *blkDispatcher) handleWithdrawalRequest(ctx context.Context, log *eth.Log, transaction *db.Db) error {
	if len(log.Data) != 32 || len(log.Topics) != 2 {
		return nil
	}
	// parse event data
	eventData := make(map[string]interface{})
	err := bld.repo.GasMonetizationAbi().UnpackIntoMap(eventData, "WithdrawalRequested", log.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack ProjectOwnerUpdated event #%d/#%d: %v", log.BlockNumber, log.Index, err)
	}
	// fetch project
	pq := transaction.ProjectQuery(ctx)
	project, err := pq.WhereProjectId(log.Topics[1].Big().Uint64()).GetFirstOrFail()
	if err != nil {
		return fmt.Errorf("failed to get project #%d: %v", log.Topics[1].Big().Uint64(), err)
	}
	// create withdrawal request
	err = transaction.StoreWithdrawalRequest(ctx, &types.WithdrawalRequest{
		ProjectId: project.Id,
		Epoch:     eventData["requestEpochNumber"].(*big.Int).Uint64(),
	})
	if err != nil {
		return fmt.Errorf("failed to store withdrawal request for project #%d: %v", project.ProjectId, err)
	}
	return nil
}
