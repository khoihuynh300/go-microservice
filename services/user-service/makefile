include .env

.PHONY: create-migration migrate-up migrate-down sqlc

create-migration:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

sqlc:
	sqlc generate
