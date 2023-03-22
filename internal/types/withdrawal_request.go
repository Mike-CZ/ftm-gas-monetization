package types

type WithdrawalRequest struct {
	Id        int64  `db:"id"`
	ProjectId uint64 `db:"project_id"`
	Epoch     uint64 `db:"epoch"`
}
