package types

import "github.com/ethereum/go-ethereum/common"

type ProjectContract struct {
	ProjectId uint64         `db:"project_id"`
	Address   common.Address `db:"address"`
}
