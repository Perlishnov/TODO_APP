.PHONY: run build test test-coverage clean wire mock swagger swagger-init swagger-clean deps help

.DEFAULT_GOAL := help

APP_NAME = todo-api
BINARY_PATH = bin/$(APP_NAME)

GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod
GOCLEAN = $(GOCMD) clean

SWAG = swag
WIRE = wire
MOCKERY = mockery

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: Run the application
run:
	$(GOCMD) run cmd/server/main.go

## build: Build binary
build:
	$(GOBUILD) -o $(BINARY_PATH) cmd/server/main.go

## test: Run unit tests
test:
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

## clean: Remove binaries and coverage files
clean:
	rm -rf bin/
	rm -f coverage.out
	$(GOCLEAN)

## wire: Generate wire_gen.go
wire:
	cd wire && $(WIRE)

## mock: Generate mocks with mockery (generates mocks for all interfaces in internal)
mock:
	$(MOCKERY) --all --dir=internal --output=mocks

## swagger-init: Generate Swagger docs
swagger-init:
	$(SWAG) init -g cmd/server/main.go -o docs --parseDependency --parseInternal

## swagger-clean: Remove generated Swagger docs
swagger-clean:
	rm -rf docs/

## swagger: Regenerate Swagger docs
swagger: swagger-clean swagger-init

## deps: Install dependencies and tools
deps:
	$(GOMOD) download
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/vektra/mockery/v2@latest
	go install github.com/swaggo/swag/cmd/swag@latest