# microservice-reference-platform

A production-oriented reference architecture for building secure, observable, independently deployable Go microservices.

> **Status:** Active development. The first milestone establishes a working HTTP vertical slice through an edge service and an orders service.

## Goals

- Keep services independently deployable.
- Use a thin edge service for authentication and routing concerns.
- Keep business logic inside domain services.
- Give each service ownership of its persistent data.
- Add cross-service infrastructure only when a real shared need emerges.
- Keep local development simple and portable to Kubernetes.

## Planned first vertical slice

Client -> Edge Service -> Orders Service -> PostgreSQL

## Repository structure

- `services/edge` - HTTP edge, reverse proxy, health/readiness, and tracing
- `services/orders` - Orders HTTP API, domain/application code, and in-memory repository
- `shared/service` - shared logging and HTTP server lifecycle
- `shared/observability` - OpenTelemetry setup and OTLP export
- `deployments/compose` - local PostgreSQL for development
- `docs` - architecture and architecture decision records
- `go.work` - local Go workspace

## Development

Copy `.env.example` to `.env`, then start the local PostgreSQL container:

```sh
docker compose -f deployments/compose/compose.yaml up -d
```

Run the services in separate terminals:

```sh
make run-edge
make run-orders
```

Edge defaults to `:8080`, Orders defaults to `:8081`, and `ORDERS_URL` must point to the Orders service.
Orders `/health` reports liveness; `/ready` checks PostgreSQL connectivity using
the `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DATABASE`, `POSTGRES_USER`, and
`POSTGRES_PASSWORD` settings. It returns 503 when PostgreSQL is unavailable.

Run formatting, license, build, test, and vet checks with:

```sh
make pre-commit
```

PostgreSQL repository integration tests run when `ORDERS_TEST_DATABASE_URL` is set
to a PostgreSQL connection string. The database user must be able to create and
drop schemas; each test uses its own temporary schema and applies the Orders migration.
Without the variable, these tests are skipped.

The Orders API is currently backed by an in-memory repository. PostgreSQL migrations are prepared in `services/orders/migrations/`; the running service is not yet wired to PostgreSQL.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
