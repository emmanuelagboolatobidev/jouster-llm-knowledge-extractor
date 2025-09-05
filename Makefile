.PHONY: help build run test clean docker-build docker-run docker-stop deps

help:
	@echo "Available commands:"
	@echo "  make deps          - Download Go dependencies"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application locally"
	@echo "  make test          - Run tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make bench         - Run benchmarks"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run with Docker Compose"
	@echo "  make docker-stop   - Stop Docker containers"

deps:
	go mod download
	go mod tidy

build: deps
	go build -o bin/api cmd/api/main.go

run: build
	./bin/api

test:
	go test -v ./...

test-verbose:
	go test -v -cover ./...

bench:
	go test -bench=. -benchmem ./...

clean:
	rm -rf bin/ data/*.db

docker-build:
	docker build -t llm-knowledge-extractor:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f

.DEFAULT_GOAL := help
