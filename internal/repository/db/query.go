package db

import (
	"context"
	"database/sql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"strings"
)

// queryBuilder is a helper to build SQL queries.
type queryBuilder[T any] struct {
	con        sqlx.ExtContext
	ctx        context.Context
	table      string
	where      []string
	parameters map[string]interface{}
	fields     string
}

// newQueryBuilder returns a new query builder.
func newQueryBuilder[T any](ctx context.Context, con sqlx.ExtContext, table string) queryBuilder[T] {
	return queryBuilder[T]{
		con:        con,
		ctx:        ctx,
		table:      table,
		fields:     "*",
		parameters: make(map[string]interface{}),
	}
}

// Select sets the fields to select.
func (qb *queryBuilder[T]) Select(fields ...string) *queryBuilder[T] {
	qb.fields = strings.Join(fields, ", ")
	return qb
}

// GetAll returns all results of the query. If no result is found, nil is returned.
func (qb *queryBuilder[T]) GetAll() ([]T, error) {
	query := qb.buildSelectQuery()
	rows, err := sqlx.NamedQueryContext(qb.ctx, qb.con, query, qb.parameters)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var result []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

// GetFirst returns the first result of the query. If no result is found, nil is returned.
func (qb *queryBuilder[T]) GetFirst() (*T, error) {
	query := qb.buildSelectQuery() + " LIMIT 1"
	rows, err := sqlx.NamedQueryContext(qb.ctx, qb.con, query, qb.parameters)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var result T
	if err := rows.StructScan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// buildSelectQuery builds the select query string.
func (qb *queryBuilder[T]) buildSelectQuery() string {
	query := "SELECT " + qb.fields + " FROM " + qb.table
	if len(qb.where) > 0 {
		query += " WHERE " + strings.Join(qb.where, " AND ")
	}
	return query
}
