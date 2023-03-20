package types

type Project struct {
	Id                  int64    `db:"id"`
	ProjectId           uint64   `db:"project_id"`
	OwnerAddress        *Address `db:"owner_address"`
	ReceiverAddress     *Address `db:"receiver_address"`
	LastWithdrawalEpoch *uint64  `db:"last_withdrawal_epoch"`
	ActiveFromEpoch     uint64   `db:"active_from_epoch"`
	ActiveToEpoch       *uint64  `db:"active_to_epoch"`
}
