APP=dredge

POSTGRES_CONTAINER=dredge-postgres
POSTGRES_IMAGE=postgres:18-alpine
POSTGRES_PORT=5432

.PHONY: gen generate mocks test lint build run docker-build tidy clean-generated frontend-build frontend-dev infra-up infra-down infra-rm

gen:
	go run github.com/ogen-go/ogen/cmd/ogen@latest --target internal/http/gen --package gen --clean api/openapi.yaml

generate: gen

mocks:
	mkdir -p internal/repository/mocks
	go run go.uber.org/mock/mockgen@latest -source "internal/repository/store.go" -destination "internal/repository/mocks/store_mock.go" -package mocks

test:
	go test ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/$(APP) ./cmd/dredge

# Build SPA and copy into internal/webui/static for go:embed (Unix shell).
frontend-build:
	cd frontend && npm ci && npm run build && rm -rf ../internal/webui/static/* && cp -r dist/* ../internal/webui/static/

frontend-dev:
	cd frontend && npm run dev

run:
	go run ./cmd/dredge

docker-build:
	docker build -t $(APP):latest .

# Local Postgres matching config.example.yaml database.dsn (postgres/postgres, db dredge).
infra-up:
	@docker info >/dev/null 2>&1 || (echo >&2 "Docker is not running or not reachable. Start Docker Desktop, then retry."; exit 1)
	@docker start $(POSTGRES_CONTAINER) 2>/dev/null || docker run -d --name $(POSTGRES_CONTAINER) -p $(POSTGRES_PORT):5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=dredge $(POSTGRES_IMAGE)
	@echo Postgres: postgres://postgres:postgres@127.0.0.1:$(POSTGRES_PORT)/dredge?sslmode=disable

infra-down:
	docker stop $(POSTGRES_CONTAINER) 2>/dev/null || true

infra-rm: infra-down
	docker rm $(POSTGRES_CONTAINER) 2>/dev/null || true

tidy:
	go mod tidy

clean-generated:
	rm -rf internal/http/gen internal/repository/mocks internal/gen
