# Herramientas de generación con versión fijada (DAS §3). Ninguna usa @latest.
SQLC_VERSION := v1.29.0
GOOSE_VERSION := v3.24.1

.PHONY: generate migracion-nueva

generate: ## Regenera internal/repository desde db/queries con sqlc
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

migracion-nueva: ## make migracion-nueva nombre=agregar_algo
	@test -n "$(nombre)" || (echo "uso: make migracion-nueva nombre=<nombre>" && exit 1)
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir db/migrations create $(nombre) sql
