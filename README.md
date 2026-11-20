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

The service and documentation structure will be introduced incrementally as the project develops.

## Development

Development prerequisites and commands will be added as the first Go modules are introduced.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
