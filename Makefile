.PHONY: build test test-race fmt vet tidy run-edge run-orders license license-check

build:
	go build ./shared/observability 
	go build ./shared/service
	go build ./services/edge/cmd
	go build ./services/orders/cmd

test:
	go test ./shared/observability/...
	go test ./shared/service/...
	go test ./services/edge/...
	go test ./services/orders/...
	
test-race:
	go test -race ./shared/observability/...
	go test -race ./shared/service/...
	go test -race ./services/edge/...
	go test -race ./services/orders/...

fmt:
	gofmt -w shared services

vet:
	go vet ./shared/observability/...
	go vet ./shared/service/...
	go vet ./services/edge/...
	go vet ./services/orders/...

tidy:
	go -C shared/observability mod tidy
	go -C shared/service mod tidy
	go -C services/edge mod tidy
	go -C services/orders mod tidy

run-edge:
	go run ./services/edge/cmd

run-orders:
	go run ./services/orders/cmd

license:
	pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/license.ps1 -Fix

license-check:
	pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/license.ps1
