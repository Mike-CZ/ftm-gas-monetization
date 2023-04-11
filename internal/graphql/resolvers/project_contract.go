package resolvers

import (
	"ftm-gas-monetization/internal/repository"
	"ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/graphql"
)

type ProjectContract struct {
	Id        graphql.Long  `db:"id"`
	ProjectId graphql.Long  `db:"project_id"`
	Address   types.Address `db:"address"`
	Approved  bool          `db:"is_approved"`
}

// Contracts provides list of contracts
func (rs *RootResolver) Contracts() (out []ProjectContract, err error) {
	query := repository.R().ProjectContractQuery()
	if err != nil {
		return nil, err
	}
	list, err := query.GetAll()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(list); i++ {
		out = append(out, ProjectContract{
			Id:        graphql.Long(list[i].Id),
			ProjectId: graphql.Long(list[i].ProjectId),
			Address:   *list[i].Address,
			Approved:  list[i].Approved,
		})
	}
	return out, nil
}
