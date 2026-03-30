.PHONY: build run run-memory run-postgres generate test lint docker-build docker-up docker-down


BUILD_DIR := build
BINARY    := $(BUILD_DIR)/ozonposts

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/app


run-memory:
	STORAGE_TYPE=memory go run ./cmd/app --config configs/config.yaml

run-postgres:
	STORAGE_TYPE=postgres go run ./cmd/app --config configs/config.yaml

run: run-memory


generate:
	go run github.com/99designs/gqlgen generate


test:
	go test -v -race -count=1 ./...

test-cover:
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html


lint:
	golangci-lint run ./...


docker-build:
	docker build -t ozonposts:latest .

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html
