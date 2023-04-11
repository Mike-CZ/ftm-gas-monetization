package resolvers

import (
	"ftm-gas-monetization/internal/repository"
	"github.com/ethereum/go-ethereum/graphql"
)

// TotalAmountClaimed provides total amount claimed tokens
func (rs *RootResolver) TotalAmountClaimed() (out graphql.Long, err error) {
	output, err := repository.R().TotalAmountClaimed()
	if err != nil {
		return 0, err
	}
	return graphql.Long(output.Int64()), err
}

// TotalAmountCollected provides total amount collected tokens
func (rs *RootResolver) TotalAmountCollected() (out graphql.Long, err error) {
	output, err := repository.R().TotalAmountCollected()
	if err != nil {
		return 0, err
	}
	return graphql.Long(output.Int64()), err
}

// TotalTransactionCount provides total amount collected tokens
func (rs *RootResolver) TotalTransactionCount() (out graphql.Long, err error) {
	output, err := repository.R().TotalTransactionsCount()
	if err != nil {
		return 0, err
	}
	return graphql.Long(output), err
}
