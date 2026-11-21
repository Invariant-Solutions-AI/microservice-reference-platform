// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package order

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)
