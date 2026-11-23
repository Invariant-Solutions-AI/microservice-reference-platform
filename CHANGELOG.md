# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2026-11-22

### Added

- Edge HTTP service for routing Orders API requests.
- Orders service with domain and application layers and PostgreSQL persistence.
- Shared structured logging and graceful HTTP server lifecycle.
- OpenTelemetry trace-context propagation and OTLP export.
- Docker Compose environment with PostgreSQL migrations and an OpenTelemetry Collector.
- API and architecture documentation, architecture decision records, and contribution guidance.
- Build, unit, race, integration, container, dependency, and vulnerability checks in CI.
- Production container images for the Edge and Orders services.

### Changed

- PostgreSQL readiness now reflects database availability.
- The initial public documentation now describes the working vertical slice and its release scope.
- PostgreSQL connections require verified TLS unless explicitly disabled for local development.
- JSON decoding rejects trailing data and Orders quantities outside the supported range.
- Go is upgraded to 1.27.3.

### Security

- PostgreSQL TLS verification is required by default.
- Container and CI checks cover service behavior and known Go vulnerabilities.
