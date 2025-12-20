include .env

.PHONY: run create-migration migrate-up migrate-down sqlc

run: 
	go run ./cmd/grpc/main.go

create-migration:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-goto:
	migrate -path migrations -database "$(DATABASE_URL)" goto $(version)

sqlc:
	sqlc generate
