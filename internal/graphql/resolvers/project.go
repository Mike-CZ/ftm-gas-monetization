package resolvers

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/graphql"
)

type Project struct {
	Id                graphql.Long  `db:"id"`
	ProjectId         graphql.Long  `db:"project_id"`
	OwnerAddress      types.Address `db:"owner_address"`
	ReceiverAddress   types.Address `db:"receiver_address"`
	CollectedRewards  graphql.Long  `db:"collected_rewards"`
	ClaimedRewards    graphql.Long  `db:"claimed_rewards"`
	RewardsToClaim    graphql.Long  `db:"rewards_to_claim"`
	Name              string        `db:"name"`
	Url               string        `db:"url"`
	ImageUrl          string        `db:"image_url"`
	TransactionsCount graphql.Long  `db:"transactions_count"`
}

// Projects provides list of projects
func (rs *RootResolver) Projects() (out []Project, err error) {
	query := repository.R().ProjectQuery()
	if err != nil {
		return nil, err
	}
	list, err := query.GetAll()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(list); i++ {
		out = append(out, Project{
			Id:                graphql.Long(list[i].Id),
			ProjectId:         graphql.Long(list[i].ProjectId),
			OwnerAddress:      *list[i].OwnerAddress,
			ReceiverAddress:   *list[i].ReceiverAddress,
			CollectedRewards:  graphql.Long(list[i].CollectedRewards.ToInt().Uint64()),
			ClaimedRewards:    graphql.Long(list[i].ClaimedRewards.ToInt().Uint64()),
			RewardsToClaim:    graphql.Long(list[i].RewardsToClaim.ToInt().Uint64()),
			Name:              list[i].Name,
			Url:               list[i].Url,
			ImageUrl:          list[i].ImageUrl,
			TransactionsCount: graphql.Long(list[i].TransactionsCount),
		})
	}
	return out, nil
}

func (pr Project) Contracts() (out []ProjectContract, err error) {
	query := repository.R().ProjectContractQuery()
	if err != nil {
		return nil, err
	}
	list, err := query.WhereProjectId(int64(pr.Id)).GetAll()
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
