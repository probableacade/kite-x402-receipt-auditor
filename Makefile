.PHONY: test run build audit

test:
	go test ./...

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -trimpath -o bin/auditor-server ./cmd/server
	go build -trimpath -o bin/auditor ./cmd/auditor

audit:
	go run ./cmd/auditor -file examples/valid-receipt.json
