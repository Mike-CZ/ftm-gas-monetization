package types

type Project struct {
	Id              uint64   `db:"id"`
	ReceiverAddress *Address `db:"receiver_address"`
}
