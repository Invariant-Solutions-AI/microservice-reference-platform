// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net/url"
	"os"
)

const defaultEdgePort = ":8080"

type Config struct {
	Addr      string
	OrdersURL *url.URL
}

func Load() (Config, error) {
	addr := os.Getenv("EDGE_ADDR")
	if addr == "" {
		addr = defaultEdgePort
	}
	value := os.Getenv("ORDERS_URL")
	if value == "" {
		return Config{}, fmt.Errorf(
			"ORDERS_URL is required",
		)
	}

	ordersURL, err := url.Parse(value)
	if err != nil {
		return Config{}, fmt.Errorf(
			"parse ORDERS_URL: %w",
			err,
		)
	}

	if ordersURL.Scheme == "" ||
		ordersURL.Host == "" {
		return Config{}, fmt.Errorf(
			"ORDERS_URL must be an absolute URL",
		)
	}

	return Config{
		Addr:      addr,
		OrdersURL: ordersURL,
	}, nil
}
