package types

type WithdrawalRequest struct {
	Id            int64   `db:"id"`
	ProjectId     int64   `db:"project_id"`
	RequestEpoch  uint64  `db:"request_epoch"`
	WithdrawEpoch *uint64 `db:"withdraw_epoch"`
	Amount        *Big    `db:"amount"`
}
