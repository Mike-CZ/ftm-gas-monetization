package db

//
//import (
//	"context"
//	"fmt"
//	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
//	"github.com/ethereum/go-ethereum/common/hexutil"
//)
//
//// AddTransaction stores a transaction reference in connected persistent storage.
//func (db *Db) AddTransaction(block *types.Block, trx *types.Transaction) error {
//	// do we have all needed data?
//	if block == nil || trx == nil {
//		return fmt.Errorf("can not add empty transaction")
//	}
//
//	// get the collection for transactions
//	col := db.client.Database(db.dbName).Collection(colTransactions)
//
//	// if the transaction already exists, we don't need to add it
//	// just make sure the transaction accounts were processed
//	if !db.shouldAddTransaction(col, trx) {
//		return db.UpdateTransaction(col, trx)
//	}
//
//	trx.LargeInput = len(trx.InputData) > trxLargeInputWall
//	if trx.LargeInput {
//		trx.InputData = hexutil.Bytes{}
//	}
//
//	if _, err := col.UpdateOne(
//		context.Background(),
//		bson.D{{Key: fiTransactionHash, Value: trx.Hash}},
//		bson.D{
//			{Key: "$set", Value: bson.D{
//				{Key: fiTransactionBlockHash, Value: trx.BlockHash},
//				{Key: fiTransactionBlockNumber, Value: trx.BlockNumber},
//				{Key: fiTransactionTimeStamp, Value: trx.TimeStamp},
//				{Key: fiTransactionFrom, Value: trx.From},
//				{Key: fiTransactionGas, Value: trx.Gas},
//				{Key: fiTransactionGasUsed, Value: trx.GasUsed},
//				{Key: fiTransactionCumulativeGasUsed, Value: trx.CumulativeGasUsed},
//				{Key: fiTransactionGasPrice, Value: trx.GasPrice},
//				{Key: fiTransactionNonce, Value: trx.Nonce},
//				{Key: fiTransactionTo, Value: trx.To},
//				{Key: fiTransactionContractAddress, Value: trx.ContractAddress},
//				{Key: fiTransactionValue, Value: trx.Value},
//				{Key: fiTransactionInputData, Value: trx.InputData},
//				{Key: fiTransactionLargeInput, Value: trx.LargeInput},
//				{Key: fiTransactionIndex, Value: trx.Index},
//				{Key: fiTransactionStatus, Value: trx.Status},
//				{Key: fiTransactionLogs, Value: trx.Logs},
//				{Key: fiTransactionOrdinalIndex, Value: trx.ComputedOrdinalIndex()},
//				{Key: fiTransactionAmount, Value: trx.ComputedAmount()},
//				{Key: fiTransactionGasGWei, Value: trx.ComputedGWei()},
//			}},
//			{Key: "$setOnInsert", Value: bson.D{
//				{Key: fiTransactionHash, Value: trx.Hash},
//			}},
//		},
//		options.Update().SetUpsert(true),
//	); err != nil {
//		db.log.Critical(err)
//		return err
//	}
//
//	// add transaction to the db
//	db.log.Debugf("transaction %s added to database", trx.Hash.String())
//
//	return nil
//}
