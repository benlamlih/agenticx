include .env
export

all: build test

build: build-openapi
	@echo "Building..."
	@go build -o agenticx cmd/agenticx/main.go

run:
	@echo "Running..."
	@go run cmd/agenticx/main.go

docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

test:
	@echo "Running tests..."
	@go test ./... -v

clean:
	@echo "Cleaning..."
	@rm -f agenticx

migrate-up:
	migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable" \
		-path ./db/migrations up

migrate-down:
	migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable" \
		-path ./db/migrations down

build-openapi:
	@echo "Bundling OpenAPI spec..."
	@mkdir -p docs/build
	npx @redocly/cli bundle docs/openapi/openapi.yaml --output docs/build/openapi.yaml
	npx @redocly/cli lint docs/build/openapi.yaml --skip-rule=no-empty-servers --skip-rule=no-server-example.com

mocks:
	@echo "Generating mocks..."
	@mockery --config .mockery.yaml

.PHONY: all build run test clean docker-run docker-down migrate-up migrate-down build-openapi mocks
