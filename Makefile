BINARY := bin/pesa
GO_FILES := $(shell find . -type f -name '*.go' -not -path './web/*')

.PHONY: up down restart logs ps dc-build dc-rebuild shell-backend shell-db psql dc-migrate seed reset dc-test \
        build run dev test test-go test-web lint fmt sec sbom clean install-tools web-build copy-dist scan scrape-crea import-crea import-bda

# ============================================================================
# Docker Compose — stack completo con un solo comando (Windows + Linux)
# ============================================================================

up:
	docker compose up -d

down:
	docker compose down

# ---- LAN deployment (http://<IP-LAN>:8080) — vedi docs/deploy-lan.md
lan-up:
	docker compose -f docker-compose.lan.yml up -d --build

lan-down:
	docker compose -f docker-compose.lan.yml down

lan-logs:
	docker compose -f docker-compose.lan.yml logs -f

restart:
	docker compose restart

logs:
	docker compose logs -f

ps:
	docker compose ps

dc-build:
	docker compose build

dc-rebuild:
	docker compose build --no-cache

shell-backend:
	docker compose exec backend sh

shell-db:
	docker compose exec db sh

psql:
	docker compose exec db psql -U pesa -d pesa

dc-migrate:
	docker compose run --rm migrate

seed:
	docker compose exec -T db psql -U pesa -d pesa < scripts/seed-foods-dev.sql

reset:
	docker compose down -v
	docker compose up -d

dc-test:
	docker compose exec backend go test ./...

# ============================================================================
# Build & tooling nativi (senza Docker)
# ============================================================================


web-build:
	cd web && npm ci && npm run build

copy-dist: web-build
	rm -rf internal/web/dist
	mkdir -p internal/web/dist
	cp -r web/dist/. internal/web/dist/

build: copy-dist
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o $(BINARY) ./cmd/pesa

run: build
	./$(BINARY) serve

dev:
	air

test: test-go test-web

test-go:
	go test ./...

test-web:
	cd web && npm test

lint:
	golangci-lint run ./...
	cd web && npx tsc --noEmit

fmt:
	gofmt -w $(GO_FILES)
	cd web && npx prettier --write "src/**/*.{ts,tsx,json,css}"

sec:
	gosec ./...
	govulncheck ./...
	cd web && npm audit --production --audit-level=high

sbom:
	cyclonedx-gomod mod -licenses -json -output sbom-go.json
	cd web && cyclonedx-npm --output-file ../sbom-web.json --output-format JSON

scan:
	go build -o bin/check-forbidden-words ./cmd/check-forbidden-words
	./bin/check-forbidden-words

install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest

clean:
	rm -rf bin tmp internal/web/dist web/dist web/node_modules

# --- Phase 2: food dataset ingestion ---

# Scrape CREA to CSV. Requires CREA outreach disposition signed (see plan-02).
scrape-crea: build
	./$(BINARY) scrape crea --output data/crea/$(shell date -u +%Y-%m-%d).csv

# Import a CREA CSV (produced by scrape-crea) into the foods table.
import-crea: build
	./$(BINARY) import crea $(CSV)

# Import a BDA IEO file (CSV or XLSX) with CREA>BDA merge precedence.
import-bda: build
	./$(BINARY) import bda $(FILE)
