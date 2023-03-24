package db

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"time"
)

func (s *DbTestSuite) TestStoreTransaction() {
	blockNumber := hexutil.Uint64(57190053)
	gasUsed := hexutil.Uint64(21000)
	big, _ := hexutil.DecodeBig("0x75bcd15")
	gasPrice := types.Big{Big: hexutil.Big(*big)}
	err := s.db.StoreTransaction(context.Background(), &types.Transaction{
		ProjectId:   1,
		Hash:        &types.Hash{Hash: common.HexToHash("0x48b50bc6f9679c37a283b308ec4cdcf14a43d818fa43e6dcbe8d9c7d28331096")},
		BlockHash:   &types.Hash{Hash: common.HexToHash("0x0002fcf20000016fb9f7ffabf8757b0d3f5f36e86bebbd09f1a03d8d4c4ae306")},
		BlockNumber: &blockNumber,
		Timestamp:   time.Unix(1678285698, 0),
		From:        &types.Address{Address: common.HexToAddress("0x391b50362bbb5adb5e0c55b120e6104363a036ab")},
		To:          &types.Address{Address: common.HexToAddress("0x391b50362bbb5adb5e0c55b120e6104363a036ab")},
		GasUsed:     &gasUsed,
		GasPrice:    &gasPrice,
	})
	assert.Nil(s.T(), err)
}
