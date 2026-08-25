.DEFAULT_GOAL := build

BINARY_NAME = flickey-api
BUILD_DIR   = ./bin

.PHONY: build run test test-race vet fmt lint tidy clean docker-build docker-up docs

## docs: Regenerate Swagger OpenAPI documentation.
docs:
	go tool swag init --generalInfo main.go --output docs --parseDependency

## build: Compile the binary.
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./...

## run: Run the server in development mode.
run:
	go run .

## test: Run all tests.
test:
	go test -v -timeout 120s ./...

## test-race: Run tests with the race detector.
test-race:
	go test -race -timeout 120s ./...

## test-integration: Run integration tests (requires running PostgreSQL + Redis).
test-integration:
	go test -v -tags integration -timeout 180s ./...

## vet: Run go vet.
vet:
	go vet ./...

## fmt: Format all Go source files.
fmt:
	gofmt -w .

## lint: Run golangci-lint (must be installed separately).
lint:
	golangci-lint run ./...

## tidy: Tidy module dependencies.
tidy:
	go mod tidy

## clean: Remove built binaries.
clean:
	rm -rf $(BUILD_DIR)

## docker-build: Build the production Docker image.
docker-build:
	docker build -t flickey-api:latest .

## docker-up: Start PostgreSQL and Redis via docker-compose.
docker-up:
	docker-compose -f ../docker-compose.yml up -d

## check: Run all quality gates (fmt, vet, test, test-race).
check: fmt vet test test-race
	@echo "All quality checks passed."
