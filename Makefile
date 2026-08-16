.PHONY: run build test test-race test-integration coverage fmt vet tidy check up up-tools down down-v logs

run:
	go run ./cmd/api

build:
	mkdir -p bin
	go build -o bin/coupons-api ./cmd/api

test:
	go test ./...

test-race:
	go test -race ./...

test-integration:
	go test -tags=integration ./internal/storage/postgres

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt vet test test-race

up:
	docker compose up --build

up-tools:
	docker compose --profile tools up --build

down:
	docker compose down

down-v:
	docker compose down -v

logs:
	docker compose logs -f api
