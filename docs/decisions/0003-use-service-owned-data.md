# ADR-0003: Use Service-Owned Data

## Status

Accepted

## Context

The Orders domain needs persistent state, while the Edge service must remain independent of Orders business data and implementation. Direct database access from a second service would couple deployment, scaling, and schema evolution.

The Orders service uses PostgreSQL for persistent storage behind its domain repository interface.

## Decision

Each domain service owns the persistent data used by its domain. No other service will read or write that data directly.

The Orders service owns its data model, persistence implementation, schema, and migrations. The other services communicate with Orders through HTTP contracts.

The in-memory implementation is a valid initial repository implementation, but it must implement the Orders domain repository interface rather than become a dependency of HTTP or domain layers.

## Current Implementation

- `services/orders/internal/persistence/memory` provides a thread-safe in-process repository.
- `services/orders/internal/domain/order.Repository` defines the persistence boundary.
- `services/orders/migrations/000001_create_orders.sql` defines the PostgreSQL schema.
- `deployments/compose/compose.yaml` starts PostgreSQL and applies migrations for local development.
- The running Orders API uses the PostgreSQL repository.

Local PostgreSQL configuration is in `.env.example`; migrations can be applied with Goose.

## Consequences

### Positive

- Domain code remains independent of database technology.
- Orders can move from memory to PostgreSQL without changing HTTP handlers or application use cases.
- Database ownership and schema evolution remain with Orders.

### Negative

- Migration and deployment coordination are required before production use.

## Guardrails

- The Edge service must not access the Orders database.
- The Orders repository interface must remain owned by the Orders domain.
- PostgreSQL-specific code must live behind that repository interface.
- Schema changes require a reviewed migration.
