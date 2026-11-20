# ADR-0001: Use Multiple Go Modules

## Status

Accepted

## Context

The repository contains multiple independently deployable services. Treating the entire repository as one Go module would make it easier for one service to import another service's implementation and would blur the intended deployment boundaries.

The project also needs a convenient local-development experience while multiple modules are being changed together.

## Decision

Each independently deployable Go service will have its own `go.mod` file.

A root `go.work` file will compose the local modules into a development workspace as those modules are introduced.

Service implementations will not import other service implementation modules. Communication between services will occur through explicit network or messaging contracts.

## Consequences

### Positive

- Module boundaries reinforce deployment boundaries.
- Each service declares only the dependencies it requires.
- Services can evolve without a repository-wide application dependency graph.
- Local development can work across modules through the workspace.

### Negative

- Dependency management is performed in more than one `go.mod` file.
- Repository-wide build and maintenance commands need to account for multiple modules.
- Shared code requires an intentional decision about whether it deserves a separate module.

## Guardrails

- Do not add a service implementation module to another service's `require` list.
- Do not create a shared module merely to remove small amounts of duplication.
- Add modules to `go.work` only when those modules actually exist.
