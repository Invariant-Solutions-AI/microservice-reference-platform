// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

const uniqueViolationCode = "23505"

type OrderRepository struct {
	pool *pgxpool.Pool
}

var _ order.Repository = (*OrderRepository)(nil)

func NewOrderRepository(
	pool *pgxpool.Pool,
) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (r *OrderRepository) Save(
	ctx context.Context,
	o *order.Order,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin save order transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO orders (
			order_id,
			customer_id,
			status,
			created_at
		)
		VALUES ($1, $2, $3, $4)
		`,
		o.OrderID,
		o.CustomerID,
		string(o.Status),
		o.CreatedAt,
	)
	if err != nil {
		return translateSaveError(
			o.OrderID,
			err,
		)
	}

	for _, item := range o.Items {
		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO order_items (
				order_id,
				product_id,
				quantity
			)
			VALUES ($1, $2, $3)
			`,
			o.OrderID,
			item.ProductID,
			item.Quantity,
		)
		if err != nil {
			return fmt.Errorf(
				"insert item for order %s: %w",
				o.OrderID,
				err,
			)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return translateSaveError(
			o.OrderID,
			err,
		)
	}

	return nil
}

func (r *OrderRepository) Get(
	ctx context.Context,
	orderID uuid.UUID,
) (*order.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var row orderRow

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
			order_id,
			customer_id,
			status,
			created_at
		FROM orders
		WHERE order_id = $1
		`,
		orderID,
	).Scan(
		&row.OrderID,
		&row.CustomerID,
		&row.Status,
		&row.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, order.ErrOrderNotFound
		}

		return nil, fmt.Errorf(
			"select order %s: %w",
			orderID,
			err,
		)
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT
			product_id,
			quantity
		FROM order_items
		WHERE order_id = $1
		ORDER BY product_id
		`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"select items for order %s: %w",
			orderID,
			err,
		)
	}
	defer rows.Close()

	itemRows := make(
		[]orderItemRow,
		0,
	)

	for rows.Next() {
		var itemRow orderItemRow

		if err := rows.Scan(
			&itemRow.ProductID,
			&itemRow.Quantity,
		); err != nil {
			return nil, fmt.Errorf(
				"scan item for order %s: %w",
				orderID,
				err,
			)
		}

		itemRows = append(
			itemRows,
			itemRow,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate items for order %s: %w",
			orderID,
			err,
		)
	}

	o, err := toDomainOrder(
		row,
		itemRows,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"map order %s: %w",
			orderID,
			err,
		)
	}

	return o, nil
}

func translateSaveError(
	orderID uuid.UUID,
	err error,
) error {
	var postgresError *pgconn.PgError

	if errors.As(err, &postgresError) &&
		postgresError.Code == uniqueViolationCode {
		return fmt.Errorf(
			"%w: %s",
			order.ErrOrderAlreadyExists,
			orderID,
		)
	}

	return fmt.Errorf(
		"save order %s: %w",
		orderID,
		err,
	)
}
