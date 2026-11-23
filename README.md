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
- `services/orders` - Orders HTTP API, domain/application code, and PostgreSQL repository
- `shared/service` - shared logging and HTTP server lifecycle
- `shared/observability` - OpenTelemetry setup and OTLP export
- `deployments/compose` - local platform and telemetry collector
- `docs` - architecture and architecture decision records
- `go.work` - local Go workspace

## Development

Start the platform, including PostgreSQL, schema migrations, and the OpenTelemetry
Collector, from the repository root:

```sh
docker compose -f deployments/compose/compose.yaml up -d --build
```

Edge is available at `http://localhost:8080` and Orders at
`http://localhost:8081`. The collector logs received traces; inspect them with
`docker compose -f deployments/compose/compose.yaml logs otel-collector`.
PostgreSQL data persists in the `postgres-data` volume across container restarts.
Stop the platform with `docker compose -f deployments/compose/compose.yaml down`;
do not use `-v` if you want to retain the data. To avoid port conflicts, set
`EDGE_PORT`, `ORDERS_PORT`, `POSTGRES_PORT`, or `OTEL_PORT` when starting Compose.

To run the Go services outside Docker instead, copy `.env.example` to `.env`,
start PostgreSQL, the migration job, and the collector:

```sh
docker compose -f deployments/compose/compose.yaml up -d postgres migrate otel-collector
```

Then run the services in separate terminals:

```sh
make run-edge
make run-orders
```

Edge defaults to `:8080`, Orders defaults to `:8081`, and `ORDERS_URL` must point to the Orders service.
Orders `/health` reports liveness; `/ready` checks PostgreSQL connectivity using
the `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DATABASE`, `POSTGRES_USER`, and
`POSTGRES_PASSWORD` settings. It returns 503 when PostgreSQL is unavailable.
New orders allow a quantity of 1 to 1,000 per item.

Orders defaults to PostgreSQL `sslmode=verify-full`, which verifies the server
certificate and hostname. Configure trusted CA certificates via `PGSSLROOTCERT`
when necessary. Local Compose and `.env.example` explicitly set
`PGSSLMODE=disable` for the development-only PostgreSQL instance; add that
setting to an existing local `.env` as needed. Do not disable TLS for remote
databases.

Run formatting, license, build, test, and vet checks with:

```sh
make pre-commit
```

PostgreSQL repository integration tests run when `ORDERS_TEST_DATABASE_URL` is set
to a PostgreSQL connection string. The database user must be able to create and
drop schemas; each test uses its own temporary schema and applies the Orders migration.
Without the variable, these tests are skipped.
The GitHub Actions build and test workflow also checks module tidiness and
workspace synchronization and runs the race and PostgreSQL integration tests.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
