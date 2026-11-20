# Architecture

## Purpose

`microservice-reference-platform` is a production-oriented reference architecture for independently deployable Go microservices behind a thin edge service.

The initial implementation deliberately starts small. The first vertical slice is:

`Client -> Edge Service -> Orders Service -> PostgreSQL`

Additional services, asynchronous messaging, shared infrastructure modules, and Kubernetes deployment are deferred until the initial service boundaries and request flow are working and tested.

## Architecture Principles

### Thin edge service

The edge service is the external entry point to the platform. Its responsibilities are limited to cross-cutting concerns that belong at the system boundary, including:

- authentication;
- request routing;
- rate limiting;
- request correlation; and
- edge-level telemetry.

The edge service must not contain business rules owned by downstream services.

### Independently deployable services

Each service is a separate Go module and deployable unit. A service owns its application behavior and may be built, tested, deployed, and scaled without importing another service's implementation.

Services communicate through explicit network or messaging contracts rather than direct Go module dependencies on other services.

### Service-owned data

A service owns the persistent data required by its domain. Other services do not access that database directly.

The initial Orders service will own its PostgreSQL schema and persistence implementation.

### Explicit contracts

Interactions between independently deployable components use explicit contracts. HTTP is the first synchronous transport. Event contracts will be introduced only when asynchronous messaging is added.

### Delay shared abstractions

Code is not moved into shared modules merely because two implementations look similar. Shared modules are introduced only after a stable cross-service infrastructure capability or integration contract emerges.

This avoids premature coupling between services.

## Initial Components

### Edge Service

Responsibilities:

- receive external HTTP requests;
- authenticate callers when authentication is introduced;
- establish request correlation;
- apply edge-level policies;
- route requests to the appropriate service; and
- expose health and readiness information.

The edge service does not own Orders business logic or Orders persistence.

### Orders Service

Responsibilities:

- implement the Orders domain;
- validate Orders business rules;
- expose the Orders HTTP API;
- persist Orders data; and
- enforce domain-specific authorization when it is introduced.

### PostgreSQL

PostgreSQL is the persistent store owned by the Orders service in the first vertical slice.

Only the Orders service accesses the Orders database directly.

## Initial Request Flow

1. A client sends an Orders request to the edge service.
2. The edge service applies boundary-level middleware and determines the destination route.
3. The edge service forwards the request to the Orders service.
4. The Orders service executes application and domain behavior.
5. The Orders service reads or writes its PostgreSQL data as required.
6. The response returns through the edge service to the client.

Authentication will be added as a later milestone. Until then, the architecture preserves the location where authentication belongs without introducing a fake production identity solution.

## Dependency Rules

Allowed dependency direction:

- service application code may depend on code within the same service module;
- infrastructure implementations may implement interfaces defined by the owning service;
- services may depend on narrowly scoped shared infrastructure modules after those modules are intentionally introduced; and
- services may exchange explicit HTTP or event contracts.

Disallowed dependency direction:

- Edge must not import the Orders implementation module.
- Orders must not import the Edge implementation module.
- One domain service must not import another domain service's implementation.
- One service must not directly query another service's database.

## Planned Repository Shape

    microservice-reference-platform/
    ├── services/
    │   ├── edge/
    │   │   ├── main.go
    │   │   ├── internal/
    │   │   └── go.mod
    │   └── orders/
    │       ├── main.go
    │       ├── internal/
    │       └── go.mod
    ├── deployments/
    ├── docs/
    │   └── decisions/
    ├── tests/
    ├── go.work
    ├── Makefile
    └── README.md

Directories are added when their corresponding implementation milestone begins rather than creating empty scaffolding up front.

## Deferred Capabilities

The following capabilities are intentionally outside the initial vertical slice:

- real OIDC/OAuth authentication;
- domain authorization;
- asynchronous messaging;
- Kafka or another event broker;
- transactional outbox;
- inventory processing;
- catalog service;
- shared observability module;
- Kubernetes manifests;
- production secret management;
- distributed tracing; and
- load and resilience testing.

They will be introduced through separate architecture decisions and commit-sized changes when the project reaches those milestones.

## Architecture Decisions

- [ADR-0001: Use Multiple Go Modules](decisions/0001-use-multiple-go-modules.md)
- [ADR-0002: Use a Thin Edge Service](decisions/0002-use-an-edge-service.md)
- [ADR-0003: Use Service-Owned Data](decisions/0003-use-service-owned-data.md)
