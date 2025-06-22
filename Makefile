include .env
export

# =================================================================================== #
# HELPERS
# =================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# =================================================================================== #
# DEVELOPMENT
# =================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## run/api/live: run the cmd/api application with live reload
.PHONY: run/api/live
run/api/live:
	@echo 'Running cmd/api with live reload...'
	wgo run ./cmd/api

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	@psql $(DB_DSN)

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
ifndef name
	$(error NAME is not set. Usage: make db/migrations/new name=your_migration_name)
endif
	@echo 'creating migration files for $(name)...'
	@migrate create -ext sql -dir ./migrations -seq ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	@migrate -path ./migrations -database $(DB_DSN) up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migrations...'
	@migrate -path ./migrations -database $(DB_DSN) down

## db/migrations/force version=$1: force a migration to a specific version
.PHONY: db/migrations/force
db/migrations/force:
	@echo 'Forcing migration to version ${version}...'
	@if [ -z "${version}" ]; then echo "Error: version is required"; exit 1; fi
	@migrate -path ./migrations -database $(DB_DSN) force ${version}

## db/migrations/reset: reset the database and apply all migrations
.PHONY: db/migrations/reset
db/migrations/reset: confirm
	@echo 'Resetting database and applying all migrations...'
	@migrate -path ./migrations -database $(DB_DSN) drop -f
	@migrate -path ./migrations -database $(DB_DSN) up

## db/migrations/goto: Go to a specific migration version (usage: make db/migrations/goto version=3)
.PHONY: db/migrations/goto
db/migrations/goto:
ifndef version
	$(error VERSION is not set. Usage: make db/migrations/goto version=3)
endif
	@echo "Migrating to version $(version)..."
	@migrate -path ./migrations -database $(DB_DSN) goto $(version)

