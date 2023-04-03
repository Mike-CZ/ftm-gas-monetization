package tracing

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
)

// TraceTransaction returns the structured logs created during the execution of
// EVM and returns them as a JSON object.
func (t *Tracer) TraceTransaction(hash common.Hash) ([]types.TransactionTrace, error) {
	var result []types.TransactionTrace
	err := t.ftm.Call(&result, "trace_transaction", hash)
	if err != nil {
		t.log.Errorf("transaction could not be traced: %s", err.Error())
		return nil, err
	}
	return result, err
}
