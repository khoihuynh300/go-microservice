include .env

.PHONY: run test test-unit test-integration test-coverage generate-mocks create-migration migrate-up migrate-down sqlc

generate-mocks:
	mockgen -package=mock_repository github.com/khoihuynh300/go-microservice/user-service/internal/repository UserRepository > mocks/repository/user_repository_mock.go
	mockgen -package=mock_repository github.com/khoihuynh300/go-microservice/user-service/internal/repository RefreshTokenRepository > mocks/repository/refresh_token_repository_mock.go
	mockgen -package=mock_repository github.com/khoihuynh300/go-microservice/user-service/internal/repository AddressRepository > mocks/repository/address_repository_mock.go
	mockgen -package=mock_password_hasher github.com/khoihuynh300/go-microservice/user-service/internal/security/password PasswordHasher > mocks/passwordhasher/password_hasher_mock.go
	mockgen -package=mock_jwt github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider JwtProvider > mocks/jwt/jwt_mock.go

run: 
	go run ./cmd/grpc/main.go

test:
	go test -v ./...

test-unit:
	go test -v -short ./tests/unit/...

test-integration:
	go test -v ./tests/integration/...

test-coverage:
	go test -coverprofile=coverage.out -coverpkg=./internal/... ./tests/...
	go tool cover -html=coverage.out

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
