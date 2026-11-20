# Contributing

Thank you for your interest in the project.

## Development principles

- Keep changes focused and reviewable.
- Prefer one architectural concern per pull request.
- Do not introduce dependencies between service implementations.
- Keep business logic inside the service that owns the domain.
- Do not add shared modules until multiple services demonstrate a stable shared need.
- Add or update tests with behavioral changes.
- Update architecture documentation when a change affects a documented design decision.

## Pull requests

A pull request should explain:

- the problem being addressed;
- the implementation approach;
- tests added or updated; and
- architectural tradeoffs when relevant.

## Security

Do not commit secrets, credentials, tokens, private keys, or customer data. See [SECURITY.md](SECURITY.md) for vulnerability-reporting guidance.

## License

By contributing to this repository, you agree that your contributions are licensed under the Apache License, Version 2.0.
