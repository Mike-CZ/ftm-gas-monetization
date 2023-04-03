// Package types implements different core types of the API.
package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TransactionTrace represents a transaction trace record.
type TransactionTrace struct {
	Action    *TransactionTraceAction `json:"action"`
	Result    *TransactionTraceResult `json:"result"`
	Subtraces int                     `json:"subtraces"`
	Error     *string                 `json:"error"`
}

type TransactionTraceAction struct {
	From *common.Address `json:"from"`
	To   *common.Address `json:"to"`
}

type TransactionTraceResult struct {
	GasUsed *hexutil.Big `json:"gasUsed"`
}
