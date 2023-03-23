package db

import (
	"context"
	"database/sql"
	"fmt"
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

// WhereId adds a where clause to the query builder.
func (qb *queryBuilder[T]) WhereId(id int64) *queryBuilder[T] {
	qb.where = append(qb.where, "id = :id")
	qb.parameters["id"] = id
	return qb
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

// GetFirstOrFail returns the first result of the query. If no result is found, an error is returned.
func (qb *queryBuilder[T]) GetFirstOrFail() (*T, error) {
	query := qb.buildSelectQuery() + " LIMIT 1"
	rows, err := sqlx.NamedQueryContext(qb.ctx, qb.con, query, qb.parameters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("no result found for query: %s", query)
	}
	var result T
	if err := rows.StructScan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFirst returns the first result of the query. If no result is found, nil is returned.
func (qb *queryBuilder[T]) GetFirst() (*T, error) {
	result, err := qb.GetFirstOrFail()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// Delete deletes all results of the query.
func (qb *queryBuilder[T]) Delete() error {
	query := "DELETE FROM " + qb.table
	if len(qb.where) > 0 {
		query += " WHERE " + strings.Join(qb.where, " AND ")
	}
	_, err := sqlx.NamedExecContext(qb.ctx, qb.con, query, qb.parameters)
	if err != nil {
		return err
	}
	return nil
}

// buildSelectQuery builds the select query string.
func (qb *queryBuilder[T]) buildSelectQuery() string {
	query := "SELECT " + qb.fields + " FROM " + qb.table
	if len(qb.where) > 0 {
		query += " WHERE " + strings.Join(qb.where, " AND ")
	}
	return query
}
