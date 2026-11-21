// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package order

import "errors"

// domain errors
var (
	ErrCustomerIDRequired = errors.New("customer ID is required")
	ErrItemsRequired = errors.New("order must contain at least one item")
	ErrProductIDRequired = errors.New("product ID is required")
	ErrInvalidQuantity = errors.New("item quantity must be greater than zero")
)

// repository errors
var (
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderNotFound      = errors.New("order not found")
)