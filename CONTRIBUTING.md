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

## Pre-commit checks

Install Go, Git, PowerShell (`pwsh`), and GNU Make, and ensure they are available
on `PATH`. On Windows, Git for Windows also provides the shell used by the hook.

## Install Goose

Install the Goose database migration CLI with Go:

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -version
```

Ensure Go's executable directory (`$(go env GOBIN)`, or `$(go env GOPATH)/bin`
when `GOBIN` is unset) is on `PATH`.

To apply the Orders migrations using the local PostgreSQL settings from
`.env.example`, run:

```sh
goose -dir services/orders/migrations postgres "host=localhost port=5432 user=orders password=local-development-only dbname=orders sslmode=disable" up
```

On Windows, first install [Chocolatey](https://chocolatey.org/install) if
`choco` is not already available:

1. Open Windows PowerShell as Administrator.
2. Review the official [installation script](https://community.chocolatey.org/install.ps1) before running it.
3. Run the following commands. The execution-policy change applies only to this PowerShell session.

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

Open a new administrator PowerShell terminal, verify Chocolatey, and install Make:

```powershell
choco --version
choco install make -y
make --version
```

Open a new terminal if necessary for the updated `PATH`.

On Debian or Ubuntu Linux:

```sh
sudo apt-get update
sudo apt-get install make
make --version
```

Install the local hook once per clone:

```sh
git config --local core.hooksPath .githooks
```

On Linux, also run `chmod +x .githooks/pre-commit`. Run `make pre-commit`
manually to perform the same checks: formatting, license headers, build,
tests, and vet. These check the working tree, including unstaged changes,
and do not modify files.

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
