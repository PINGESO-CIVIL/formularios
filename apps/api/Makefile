# Comandos del backend. Ver AGENTS.md en la raÃÆ’Æ’Ã‚Â­z.
SHELL := /bin/bash

ENV_FILE ?= ../../.env
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export
endif

DATABASE_URL ?= postgres://formularios:formularios@localhost:5432/formularios?sslmode=disable
MIGRATIONS_DIR := db/migrations

GOLANGCI_LINT ?= golangci-lint
SQLC ?= sqlc
GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@v3.24.1
TRIVY ?= trivy

.PHONY: run worker build test test-int lint trivy sqlc migrate-up migrate-down migrate-new check tools tidy

run: ## API en modo dev (requiere docker compose up -d)
	APP_ENV=$${APP_ENV:-dev} go run ./cmd/api

worker: ## Worker (outbox relay)
	APP_ENV=$${APP_ENV:-dev} go run ./cmd/worker

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

test: ## Pruebas unitarias + coverage.out
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-int: ## Pruebas de integraciÃÆ’Æ’Ã†â€™ÃÆ’â€Å¡Ã‚Â³n (Postgres en contenedor vÃÆ’Æ’Ã†â€™ÃÆ’â€Å¡Ã‚Â­a testcontainers)
	go test -race -tags integration ./e2e/...

lint:
	$(GOLANGCI_LINT) run ./...

trivy: ## Escaneo de vulnerabilidades en dependencias y Dockerfile
	$(TRIVY) fs --scanners vuln,secret,misconfig --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 .

sqlc:
	$(SQLC) generate

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down: ## Revierte la ÃÆ’Æ’Ã†â€™ÃÆ’â€Å¡Ã‚Âºltima migraciÃÆ’Æ’Ã†â€™ÃÆ’â€Å¡Ã‚Â³n
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-new: ## make migrate-new name=create_form_templates
	@test -n "$(name)" || (echo "uso: make migrate-new name=<nombre>" && exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) -s create $(name) sql

check: lint test test-int ## CÃÆ’Æ’Ã†â€™ÃÆ’â€Å¡Ã‚Â³rrelo antes de terminar una tarea

tidy:
	go mod tidy

tools: ## Instala herramientas de desarrollo
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
