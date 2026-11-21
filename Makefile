.PHONY: build test fmt vet tidy run-edge run-orders

build:
	go build ./services/edge/cmd ./services/orders/cmd

test:
	go test ./services/edge/... ./services/orders/...

fmt:
	go fmt ./services/edge/... ./services/orders/...

vet:
	go vet ./services/edge/... ./services/orders/...

tidy:
	go -C services/edge mod tidy
	go -C services/orders mod tidy

run-edge:
	go -C services/edge run ./cmd

run-orders:
	go -C services/orders run ./cmd
