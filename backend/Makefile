# Comandos del backend. Ver backend/AGENTS.md.
SHELL := /bin/bash

ENV_FILE ?= ../.env
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export
endif

DATABASE_URL ?= postgres://formularios:formularios@localhost:5432/formularios?sslmode=disable
MIGRATIONS_DIR := db/migrations

GOLANGCI_LINT ?= golangci-lint
SQLC ?= sqlc
MIGRATE ?= migrate

.PHONY: run worker build test test-int lint sqlc migrate-up migrate-down migrate-new check tools tidy

run: ## API en modo dev (requiere docker compose up -d)
	APP_ENV=$${APP_ENV:-dev} go run ./cmd/api

worker: ## Worker (outbox relay)
	APP_ENV=$${APP_ENV:-dev} go run ./cmd/worker

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

test: ## Pruebas unitarias
	go test -race ./...

test-int: ## Pruebas de integración (Postgres en contenedor vía testcontainers)
	go test -race -tags integration ./test/integration/...

lint:
	$(GOLANGCI_LINT) run ./...

sqlc:
	$(SQLC) generate

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down: ## Revierte la última migración
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-new: ## make migrate-new name=create_form_templates
	@test -n "$(name)" || (echo "uso: make migrate-new name=<nombre>" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq -digits 4 $(name)

check: lint test test-int ## Córrelo antes de terminar una tarea

tidy:
	go mod tidy

tools: ## Instala herramientas de desarrollo
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
