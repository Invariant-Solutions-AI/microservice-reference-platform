// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package orders

import (
	"time"

	"github.com/google/uuid"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/orders/internal/domain/order"
)

type Create struct {
	repository order.Repository
	newID func() uuid.UUID
	now func() time.Time
}

func NewCreate(
	repository order.Repository,
	newID func() uuid.UUID,
	now func() time.Time,
) *Create {
	return &Create{
		repository: repository,
		newID:      newID,
		now:        now,
	}
}