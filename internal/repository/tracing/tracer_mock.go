package tracing

import (
	"ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

// TracerMock is mock of the TracingRpc client.
type TracerMock struct {
	mock.Mock
}

func (tm *TracerMock) TraceTransaction(hash common.Hash) ([]types.TransactionTrace, error) {
	args := tm.Called(hash)
	return args.Get(0).([]types.TransactionTrace), args.Error(1)
}
