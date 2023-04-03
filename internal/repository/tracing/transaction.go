package tracing

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

// TraceTransaction returns the structured logs created during the execution of
// EVM and returns them as a JSON object.
func (t *Tracer) TraceTransaction(hash common.Hash) {
	traceOpts := map[string]string{
		"tracer":  "callTracers",
		"timeout": "119s",
	}
	var result interface{}
	err := t.ftm.Call(&result, "debug_traceTransaction", hash, traceOpts)
	if err != nil {
		t.log.Errorf("transaction could not be traced: %s", err.Error())
	}
	panic(fmt.Sprintf("%+v", result))
}
