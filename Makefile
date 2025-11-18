.PHONY: \
	generate-oapi generate-repo-mocks init \
	run \
	docker-up docker-down docker-down-hard docker-logs docker-rebuild \
	test-repo test-unit test-e2e test-load e2e

ENV_FILE := .env

HAS_OLD_COMPOSE := $(shell command -v docker-compose >/dev/null 2>&1 && echo 1 || echo 0)

ifeq ($(HAS_OLD_COMPOSE),1)
	DC := docker-compose
else
	DC := docker compose
endif

GENERATED_OAPI_DIR := internal/api
OAPI_PKG := github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

generate-oapi:
	@command -v oapi-codegen >/dev/null 2>&1 || go install $(OAPI_PKG)
	@mkdir -p $(GENERATED_OAPI_DIR)
	@oapi-codegen -generate types      -package api -o $(GENERATED_OAPI_DIR)/types.gen.go  api/openapi.yaml
	@oapi-codegen -generate chi-server -package api -o $(GENERATED_OAPI_DIR)/server.gen.go api/openapi.yaml

INTERFACES_REPO := PRRepository UserRepository TeamRepository
MOCK_DIR        := mocks
REPO_DIR        := internal/repo/postgres

generate-repo-mocks:
	@mkdir -p $(MOCK_DIR)
	go install github.com/vektra/mockery/v2@latest
	@for iface in $(INTERFACES_REPO); do \
		mockery --name=$$iface --tags mockery --dir=$(REPO_DIR) --output=$(MOCK_DIR); \
	done

init: generate-oapi generate-repo-mocks

run: docker-up

docker-up:
	$(DC) up --build

docker-down:
	$(DC) down

docker-down-hard:
	$(DC) down -v

docker-rebuild:
	$(DC) build --no-cache

docker-logs:
	$(DC) logs -f app

test-repo:
	$(DC) up -d db
	MIGRATIONS_DIR="$(PWD)/migrations" \
	go test -p 1 ./internal/repo/postgres/test/...

test-unit:
	@go test ./internal/...

test-e2e:
	@set -e; \
	$(DC) -f docker-compose.e2e.yml up -d --build app; \
	status=0; \
	go test ./e2e -v -count=1 || status=$$?; \
	$(DC) -f docker-compose.e2e.yml down -v; \
	exit $$status

# k6 required
test-load:
	$(DC) up -d
	k6 run k6/main.js
