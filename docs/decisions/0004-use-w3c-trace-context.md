# ADR-0004: Use W3C Trace Context for Distributed Request Correlation

## Status

Accepted

## Context

Requests to the microservice reference platform may cross multiple independently deployable components.

The initial synchronous request flow is:

`Client -> Edge Service -> Orders Service`

Future architecture may introduce additional services, asynchronous messaging, background workers, and other distributed processing.

Troubleshooting these flows requires a way to correlate activity across component boundaries without relying on timestamps or service-specific identifiers.

Several approaches were considered:

1. A custom request identifier such as `X-Request-ID`
2. Microsoft Correlation Vector
3. W3C Trace Context

The selected mechanism should:

- Propagate across service boundaries.
- Support distributed tracing rather than only log correlation.
- Avoid introducing a proprietary protocol where an industry standard exists.
- Support future observability tooling without changing the application's correlation model.
- Work across languages and service implementations.
- Allow services to participate in a trace without depending on each other's implementations.

## Decision

The platform will use **W3C Trace Context** as its primary mechanism for propagating distributed trace identity between services.

HTTP services will use the standard `traceparent` and, when required, `tracestate` headers rather than defining an application-specific correlation header.

The Edge service is the first platform boundary responsible for participating in trace-context propagation.

Downstream services will participate in the same trace context when processing requests.

Application logging and future distributed tracing instrumentation should obtain correlation information from the active trace context rather than establishing an independent correlation identifier.

The application will not implement its own W3C Trace Context parser when supported instrumentation or tracing libraries can manage propagation correctly.

## Rationale

W3C Trace Context defines standard HTTP headers and formats for propagating context information used in distributed tracing.

The specification provides a common format that can be understood across different tracing systems and service implementations.

This aligns with the goals of the reference platform because service boundaries should remain independent of:

- Programming language
- Tracing vendor
- Logging implementation
- Deployment environment
- Observability backend

Using a standard trace context also allows correlation and distributed tracing to share the same underlying identity rather than maintaining separate request-ID and trace-ID systems.

## Trace Model

A request entering the platform participates in a distributed trace.

Conceptually:

`Client -> Edge -> Orders`

Each participating component processes work within that trace.

As the architecture evolves, a trace may span additional components:

`Client -> Edge -> Orders -> Downstream Service`

The trace identifies the distributed operation while tracing instrumentation can represent individual operations within that trace.

The application should not treat the distributed trace identifier as a business identifier.

For example, an `OrderID` continues to identify an order regardless of which request or trace created or retrieved it.

## HTTP Propagation

The platform will use the standard W3C HTTP trace-context headers:

- `traceparent`
- `tracestate`

`traceparent` carries the portable trace context used between participating systems.

`tracestate` may carry additional tracing-system-specific state when required.

Application handlers should not manually rewrite these headers unless required by the tracing implementation.

Trace propagation should eventually be performed by the platform's tracing instrumentation.

## Middleware

Cross-cutting HTTP behavior related to trace context belongs in middleware rather than individual route handlers.

Conceptually:

`Request -> Trace Context Middleware -> Logging Middleware -> Routing -> Handler`

This keeps application endpoints independent of tracing mechanics.

For example, an Orders handler should not contain logic whose primary responsibility is reading or creating trace headers.

The Edge service will establish the initial middleware integration. The same architectural pattern can later be applied to downstream services.

## Logging

Structured logs should include trace identifiers when an active trace context exists.

This allows operators to move between logs and distributed traces without introducing a second unrelated correlation mechanism.

Logging must not assume that a trace identifier identifies a user, order, customer, or other domain entity.

Domain identifiers should be logged separately when appropriate.

## Alternative: Custom Request ID

A simpler implementation would generate an identifier at the Edge and propagate it using a header such as:

`X-Request-ID`

### Advantages

- Simple to implement
- Simple to understand
- Sufficient for basic log correlation

### Disadvantages

- Defines an application-specific convention
- Represents correlation separately from distributed tracing
- Requires decisions about generation, validation, propagation, and interoperability
- Risks maintaining both request IDs and trace IDs after tracing is introduced

### Decision

Rejected as the primary correlation mechanism.

A separate request ID may be introduced in the future only if a requirement emerges that is not satisfied by trace context.

## Alternative: Microsoft Correlation Vector

Microsoft Correlation Vector is a protocol for tracing and correlating events through distributed systems using a lightweight vector-clock model.

Its design includes causal ordering capabilities beyond a simple request identifier.

Correlation Vector is particularly interesting for scenarios where causality and event ordering need to be represented without relying solely on system-clock time.

### Advantages

- Represents causal relationships in distributed activity
- Supports ordering scenarios beyond a simple request identifier
- Has existing implementations across multiple languages
- Provides an established model rather than an ad hoc correlation identifier

### Disadvantages

- Adds a correlation protocol in addition to the broader distributed tracing ecosystem
- Requires Correlation Vector-specific processing and propagation
- Introduces additional protocol concepts compared with standard trace-context propagation
- The public Microsoft CorrelationVector repository is archived and read-only

### Decision

Rejected as the primary correlation mechanism for this reference platform.

Correlation Vector remains a useful architectural reference for understanding causal tracing and may be appropriate in systems with requirements specifically aligned with vector-based correlation.

The reference platform instead prioritizes standardized cross-platform trace propagation.

## Consequences

### Positive

- Correlation uses an established distributed tracing standard.
- Services remain independent of a particular observability vendor.
- The architecture can adopt distributed tracing without replacing an earlier custom request-ID model.
- Logging and tracing can share trace identity.
- Trace propagation can cross language and service boundaries.
- Edge-specific correlation code is minimized.

### Negative

- Trace context is more sophisticated than a simple generated request identifier.
- Correct trace participation requires understanding trace and operation boundaries.
- Observability instrumentation becomes part of the platform architecture.
- Developers must avoid treating trace identifiers as domain identifiers.

## Guardrails

- Do not create a new correlation header merely for convenience.
- Do not generate independent request IDs when an existing valid trace context already provides the required correlation.
- Do not use trace IDs as `OrderID`, `CustomerID`, or any other domain identifier.
- Do not encode business or personally identifiable information into trace-context values.
- Do not manually implement W3C trace-context parsing when supported instrumentation can provide standards-compliant propagation.
- Keep tracing concerns outside domain and application packages.
- Prefer HTTP middleware and transport instrumentation for inbound and outbound propagation.
- Structured logs should include trace identity when available but must remain useful when tracing is unavailable.

## Implementation Sequence

The implementation will be introduced incrementally.

### Phase 1: Trace Context Propagation

Add HTTP middleware and propagation support at the Edge boundary.

Verify that trace context survives:

`Client -> Edge -> Orders`

### Phase 2: Structured Logging

Include active trace identifiers in structured request logs.

### Phase 3: Distributed Tracing

Introduce tracing instrumentation and spans while continuing to use the same W3C trace context.

### Phase 4: Asynchronous Propagation

When asynchronous messaging is introduced, define how trace context crosses message boundaries without coupling domain event contracts to tracing implementation details.

## References

- W3C Trace Context Recommendation: https://www.w3.org/TR/trace-context/
- W3C Trace Context HTTP request-header format: https://github.com/w3c/trace-context/blob/main/spec/20-http_request_header_format.md
- Microsoft Correlation Vector: https://github.com/microsoft/CorrelationVector
- Microsoft Correlation Vector v2.1: https://github.com/microsoft/CorrelationVector/blob/master/cV%20-%202.1.md