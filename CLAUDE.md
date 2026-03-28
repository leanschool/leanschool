# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Leanschool is a school management system with microservices in Go and a React frontend. Services:
- **leanschool** — Main backend (receipts, accounts, school classes, users) with PostgreSQL + Keycloak
- **receipt-reader** — OCR service using Tesseract + ImageMagick for receipt image processing
- **file-service** — File handling microservice
- **leanschool-model** — Shared Go module with common models and OpenAPI spec generator
- **leanschool-ui** — React + Vite frontend with Capacitor (iOS/Android) and PWA support

## Commands

### Backend (root Makefile)
```bash
make build                  # Build leanschool and receipt-reader binaries to bin/
make build-leanschool       # Build leanschool only
make build-receipt-reader   # Build receipt-reader only
make tidy                   # Run go mod tidy for all modules
make spec                   # Generate OpenAPI specs to docs/
make publish-model          # Publish leanschool-model as git tag
make clean                  # Remove bin/
```

### Frontend (`leanschool-ui/`)
```bash
npm run dev       # Vite dev server
npm run build     # Production build
npm run lint      # ESLint
npm run preview   # Preview production build
```

### Go tests (run from service directory)
```bash
cd leanschool && go test ./...
cd receipt-reader && go test ./...
```

## Architecture

### Backend Services
Each Go service follows the same layout:
- `cmd/main.go` — Entry point: HTTP server, DB connection, Keycloak middleware setup
- `internal/handler/` — HTTP handlers + CORS/auth middleware
- `internal/storage/` — PostgreSQL storage layer

Authentication is handled by Keycloak (OIDC). The `leanschool` service validates JWTs from Keycloak using a client secret. Handlers check roles from JWT claims.

The `leanschool-model` module is a separate Go module imported by services via `go.mod` replace directives or tagged versions. After modifying models, run `make publish-model` to tag, then update dependent services.

The `receipt-reader` service runs a multi-step OCR pipeline: image upload → ImageMagick preprocessing → Tesseract OCR → structured extraction. It fetches/stores files via the `file-service`.

### Frontend (`leanschool-ui/src/`)
- `auth/` — OIDC/PKCE authentication utilities (integrates with Keycloak)
- `components/` — React UI screens and forms
- `i18n/` — Internationalization/translations
- Environment variables are prefixed `VITE_` (see `.env.example`)

### Infrastructure (`infra/`)
- `dockerfiles/` — Multi-stage Dockerfiles for each service
- `k3d/` — Kubernetes manifests for local dev cluster (Keycloak, PostgreSQL, OpenBao, service deployments)
- `local/` — Local development configs

### Environment Variables
**leanschool**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `KEYCLOAK_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`, `KEYCLOAK_CLIENT_SECRET`, `RECEIPT_TEMPLATE_PATH`

**receipt-reader**: `TESSERACT_LANG` (default: `eng`), `FILE_SERVICE_URL`

**leanschool-ui**: `VITE_RECEIPT_READER_URL`, `VITE_LEANSCHOOL_URL`, `VITE_FILE_SERVICE_URL`, `VITE_KEYCLOAK_URL`, `VITE_KEYCLOAK_REALM`, `VITE_KEYCLOAK_CLIENT_ID`

## CI/CD
GitLab CI (`gitlab-ci.yaml`) runs three stages: build Docker images → publish to registry → deploy (manual, not yet implemented).

## Rule files
Implementation task rules live in `.claude/rules/`. Each rule file describes a goal, the tasks needed to reach it, and a running log of completed work at the bottom (so future sessions can pick up where the last one left off).

Current rules:
- `datamodel.md` — domain model implementation (Go backend: models, storage, handlers)
- `ui_components.md` — reusable UI components for each domain model
- `security.md` — security-related tasks
