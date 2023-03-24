package types

type ProjectContract struct {
	Id        int64    `db:"id"`
	ProjectId int64    `db:"project_id"`
	Address   *Address `db:"address"`
	Enabled   bool     `db:"is_enabled"`
}
