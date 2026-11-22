// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open creates and verifies a PostgreSQL connection pool.
func Open(
	ctx context.Context,
	connectionString string,
) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf(
			"parse PostgreSQL connection string: %w",
			err,
		)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf(
			"create PostgreSQL connection pool: %w",
			err,
		)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"pinging PostgreSQL: %w",
			err,
		)
	}

	return pool, nil
}
