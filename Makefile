APP_ENTRY := ./cmd/api
MIGRATE_ENTRY := ./cmd/migrate

.PHONY: fmt test run migrate-up migrate-down migrate-version

fmt:
	go fmt ./...

test:
	go test ./...

run:
	go run $(APP_ENTRY)

migrate-up:
	go run $(MIGRATE_ENTRY) up

migrate-down:
	go run $(MIGRATE_ENTRY) down

migrate-version:
	go run $(MIGRATE_ENTRY) version
