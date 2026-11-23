# microservice-reference-platform

A reference implementation for building independently deployable Go microservices
behind a thin HTTP edge service.

> **Release scope:** The initial public release demonstrates an early working
> vertical slice; this is not a production-ready application or deployment
> template. Review the [architecture](docs/architecture.md) and [security
> policy](SECURITY.md) before adapting it.

## What is included

- An HTTP edge service that routes Orders requests.
- An Orders service with domain and application layers and PostgreSQL persistence.
- Shared structured logging, graceful HTTP server lifecycle, and OpenTelemetry
  trace-context propagation and OTLP export.
- A local Docker Compose environment with PostgreSQL migrations and an
  OpenTelemetry Collector.
- Go unit and integration tests and CI checks for build quality, dependencies,
  known Go vulnerabilities, and container behavior.

Authentication and authorization, asynchronous messaging, Kubernetes
deployment, and production secret management are not implemented. Do not expose
the sample services to untrusted networks or use the Compose credentials in a
deployed environment. See [Deferred Capabilities](docs/architecture.md#deferred-capabilities)
for the current scope.

## Run locally

You need Docker Engine with Docker Compose v2. From the repository root, build
and start the platform:

```sh
docker compose -f deployments/compose/compose.yaml up -d --build
```

The edge service is available at `http://localhost:8080`; the Orders service is
also exposed directly at `http://localhost:8081` for local development. Compose
binds these ports to loopback. It starts PostgreSQL, applies the Orders schema
migration, and starts the OpenTelemetry Collector. The collector logs received
traces; no tracing backend is included.

Check service health and readiness:

```sh
curl -i http://localhost:8080/health
curl -i http://localhost:8081/ready
```

Create and retrieve an order through the edge API:

```sh
curl -i -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customerId": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
    "items": [
      {
        "productId": "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
        "quantity": 2
      }
    ]
  }'
```

The response includes the generated `orderId`; use it to retrieve the order:

```sh
curl -i http://localhost:8080/api/v1/orders/REPLACE_WITH_ORDER_ID
```

See the [Orders API reference](docs/api.md) for routes, request validation, and
error responses.

PostgreSQL data persists in the `postgres-data` volume across restarts. Stop
the platform with:

```sh
docker compose -f deployments/compose/compose.yaml down
```

This keeps the database volume. Add `--volumes` only when you intend to delete
the local database. To change the host ports, set `EDGE_PORT`, `ORDERS_PORT`,
`POSTGRES_PORT`, or `OTEL_PORT` before running Compose.

## Run services outside Docker

Go `1.27.4` is required. If you do not already have a local `.env`, copy
`.env.example` to `.env` in the repository root. The service binaries load this
file when started from the root. Start the database, migration job, and collector:

```sh
docker compose -f deployments/compose/compose.yaml up -d postgres migrate otel-collector
```

In separate terminals, run:

```sh
make run-edge
make run-orders
```

The Orders service defaults to PostgreSQL TLS verification
(`PGSSLMODE=verify-full`). The example `.env` sets `PGSSLMODE=disable` only for
the local Compose database. Keep TLS verification enabled for remote databases.
The sample database password is for local development only.

## Development and CI

Install Go `1.27.4`, Git, GNU Make, PowerShell (`pwsh`), and Docker Compose v2
for the full local checks. Run:

```sh
make pre-commit
```

This runs formatting and license checks, module tidying, vet, builds, and unit
tests. The module-tidying step can update `go.mod` or `go.sum`; review any such
changes. PostgreSQL integration tests require `ORDERS_TEST_DATABASE_URL` and a
database user that can create and drop schemas.

GitHub Actions also runs race tests, PostgreSQL integration tests, container
smoke tests (including persistence across an Orders restart), dependency
review on pull requests, and `govulncheck` across all Go modules.

See [Contributing](CONTRIBUTING.md) for contributor guidelines and local hook
setup.

## Repository structure

- `services/edge` - external HTTP edge, routing, health/readiness, and tracing
- `services/orders` - Orders API, domain/application code, PostgreSQL repository,
  and migrations
- `shared/service` - shared logging and HTTP server lifecycle
- `shared/observability` - OpenTelemetry setup and OTLP export
- `deployments/compose` - local services and telemetry collector
- `docs` - architecture, API reference, and architecture decision records
- `go.work` - local Go workspace

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

Release history is documented in [CHANGELOG.md](CHANGELOG.md).
