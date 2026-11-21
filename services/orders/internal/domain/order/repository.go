// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package order

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, order *Order) error
	Get(ctx context.Context, id uuid.UUID) (*Order, error)
}