# https://tech.davis-hansson.com/p/make/
SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build            Build the project"
	@echo "  clean            Remove build artifacts"
	@echo "  format           Format the source code"
	@echo "  lint             Lint the source code"
	@echo "  test             Run tests"
	@echo "  test-coverage    Run tests with coverage"
	@echo "  test-coverage-report Run tests with coverage and open HTML report"
	@echo "  test-coverage-func   Run tests with function coverage details"

.PHONY: format
format:
	go fmt ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	go test -v -cover ./...

.PHONY: test-coverage-report
test-coverage-report:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

.PHONY: test-coverage-func
test-coverage-func:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

