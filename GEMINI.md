# Project: Citual (Modular SaaS)

## 1. Project Overview
**Citual** is a modular SaaS platform. The backend follows a **Modular Monolith** pattern using **Hexagonal Architecture** (Ports and Adapters). The frontend is a separate Next.js application.

* **Repository:** `github.com/ranakdinesh/citual`
* **Primary Goal:** Build a scalable, module-driven SaaS where each module (Identity, CRM, CMS, etc.) acts as a self-contained domain.
* **Key Constraint:** Shared Platform services (DB, Logger, Temporal) are injected into modules. Modules must NOT depend heavily on each other; they communicate via defined interfaces (Ports).

## 2. Tech Stack

### Backend (Golang)
* **Language:** Go (Golang)
* **Router:** `go-chi/chi`
* **Database:** PostgreSQL
* **Drivers:** `pgx/v5` (using `pgxpool` for connection pooling)
* **ORM/DAO:** `sqlc` (Type-safe SQL generation). **NO GORM.**
* **Architecture:** Hexagonal (Ports & Adapters).
* **Orchestration:** **Temporal** (For long-running workflows and Python AI Agent orchestration).
* **Structure:**
    * `internal/platform`: Shared tech (Config, Logger, DB Pool, Temporal Client).
    * `internal/modules`: Business Domains (Identity, CRM, CMS, etc.).

### Frontend (Next.js)
* **Framework:** Next.js (App Router preferred)
* **Language:** TypeScript
* **Styling:** Tailwind CSS
* **Source:** References templates in `./design` folder.

### Infrastructure & CI/CD
* **Containerization:** Docker & Docker Compose (`docker-compose.itest.yml`).
* **CI/CD:** GitHub Actions.
* **Feedback Loop:** Automated testing + Vulnerability Assessment.

## 3. Directory Structure Rules
The backend structure is **strict**. Do not invent new top-level folders.

```text
├── cmd/api/                  # Main entrypoint. Wires App + Modules.
├── internal/
│   ├── app/                  # Application Composition Root.
│   │   ├── app.go            # Initializes Platform & Modules.
│   │   └── router.go         # Main Chi router mounting module sub-routers.
│   ├── platform/             # SHARED INFRASTRUCTURE (Adapters only).
│   │   ├── temporal/         # Temporal Client Wrapper.
│   │   └── postgres/         # Shared pgxpool instance.
│   └── modules/              # DOMAIN MODULES (Hexagonal).
│       ├── identity/
│       ├── crm/
│       ├── cms/
│       ├── communication/
│       ├── social/
│       ├── tracker/
│       └── support/
```
## 4. Module Layout (Hexagonal Standards)
Every module inside internal/modules/<name>/ MUST follow this internal structure:

module.go: The ONLY entry point. It accepts platform dependencies and returns a Module struct with the public HTTP router.

core/domain/: Pure Go structs. No SQL tags, no JSON tags (unless necessary). STRICTLY NO IMPORTS from adapters.

core/ports/: Interfaces defining HOW the domain interacts with the world (e.g., UserRepository, EmailService).

core/services/: Business logic implementing the use cases.

adapters/postgres/: Implementation of repository ports using sqlc + pgx.

adapters/http/: HTTP Handlers, DTOs, and Routing specific to this module.

## 5. Coding Standards & Guidelines
### Backend (Go)

# Project: Citual (Identity Module)

## 1. Architecture Guardrails (STRICT)
* **Pattern:** Hexagonal Architecture (Ports & Adapters).
* **Database:**
  * **Driver:** `pgx/v5` (via `pgxpool`).
  * **ORM:** `sqlc` (Type-safe, generated code). **NO GORM.**
  * **Transaction Management:** MUST use `internal/platform/db.RunInTx`.
  * **Store Pattern:** ONE single `postgres.Store` struct that wraps the `pgxpool` and `sqlc.Queries`. Do NOT create separate DB connections for Fosite.
* **Tenancy:**
  * **Hard Constraint:** Every domain query MUST filter by `tenant_id`.
  * **RLS:** The `db.RunInTx` helper handles RLS context injection (`set_config`).
* **Identity Provider:** Ory Fosite v0.49+.

## 2. Package Structure & Imports
* **Root:** `github.com/ranakdinesh/citual`
* **Platform (Shared):** `internal/platform/...`
  * DB Helper: `internal/platform/db` (Contains `RunInTx`, `GetTx`).
* **Identity Module:** `internal/modules/identity/...`
  * **Domain:** `core/domain` (Pure Go structs).
  * **Ports:** `core/ports` (Interfaces).
  * **Postgres Adapter:** `adapters/postgres` (Holds `Store` struct and `sqlc` generated code).
  * **Fosite Adapter:** `adapters/fosite_store` (Implements `fosite.Storage` using `postgres.Store`).

## 3. The `postgres.Store` Contract
The `adapters/postgres` package exports a `Store` struct.
```go
type Store struct {
    Pool    *pgxpool.Pool
    Queries *sqlc.Queries
}
// Helper to get queries with active transaction if present
func (s *Store) getQueries(ctx context.Context) *sqlc.Queries { ... }