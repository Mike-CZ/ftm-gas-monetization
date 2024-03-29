package types

type Project struct {
	Id                  int64    `db:"id"`
	ProjectId           uint64   `db:"project_id"`
	OwnerAddress        *Address `db:"owner_address"`
	ReceiverAddress     *Address `db:"receiver_address"`
	Name                string   `db:"name"`
	Url                 string   `db:"url"`
	ImageUrl            string   `db:"image_url"`
	LastWithdrawalEpoch *uint64  `db:"last_withdrawal_epoch"`
	CollectedRewards    *Big     `db:"collected_rewards"`
	ClaimedRewards      *Big     `db:"claimed_rewards"`
	RewardsToClaim      *Big     `db:"rewards_to_claim"`
	TransactionsCount   uint64   `db:"transactions_count"`
	ActiveFromEpoch     uint64   `db:"active_from_epoch"`
	ActiveToEpoch       *uint64  `db:"active_to_epoch"`
}
