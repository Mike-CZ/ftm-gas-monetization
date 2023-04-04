// Package types implements different core types of the API.
package types

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
)

// TransactionTrace represents a transaction trace record.
type TransactionTrace struct {
	Action       *TransactionTraceAction `json:"action"`
	Result       *TransactionTraceResult `json:"result"`
	Error        *string                 `json:"error"`
	TraceAddress []int                   `json:"traceAddress"`
}

// StringPath returns the trace address as a string path.
func (t *TransactionTrace) StringPath() string {
	return strings.Trim(strings.Replace(fmt.Sprint(t.TraceAddress), " ", "", -1), "[]")
}

// ParentStringPath returns the trace address of a prent as a string path.
func (t *TransactionTrace) ParentStringPath() *string {
	path := t.StringPath()
	// if the path is empty, there is no parent
	if len(path) == 0 {
		return nil
	}
	// remove the last element
	path = path[:len(path)-1]
	return &path
}

type TransactionTraceAction struct {
	From *common.Address `json:"from"`
	To   *common.Address `json:"to"`
}

type TransactionTraceResult struct {
	GasUsed *hexutil.Uint64 `json:"gasUsed"`
}
