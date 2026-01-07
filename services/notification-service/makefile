include .env

.PHONY: run test test-unit test-integration test-coverage

run: 
	go run ./cmd/server/main.go

test:
	go test -v ./...

test-unit:
	go test -v -short ./tests/unit/...

test-integration:
	go test -v ./tests/integration/...

test-coverage:
	go test -coverprofile=coverage.out -coverpkg=./internal/... ./tests/...
	go tool cover -html=coverage.out
