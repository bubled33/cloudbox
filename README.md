# Cloud File Storage (Google Drive Lite)

> **Language / Язык**: **EN** | [RU](./docs/README_RU.md)

---

### Introduction

**Cloud File Storage** is a Go‑based web service for storing, versioning, and sharing files via short‑lived public links, with background workers for preview generation and system metrics.  
The project follows a layered, hexagonal architecture with explicit domain, application, infrastructure, API, and worker layers to keep business logic isolated from frameworks and external systems.  

**Key features:**  
- Passwordless authentication via email magic links and session tokens.  
- File versioning with restore, metadata updates, and soft constraints on size and capacity.  
- Temporary public links with strict TTL and automatic expiration workers.  
- Asynchronous preview generation using a queue and dedicated worker.  
- Built‑in metrics endpoint and OpenTelemetry tracing for observability.  
- OpenAPI/Swagger documentation served directly from the running API.  

---

### Non‑Functional Requirements

The system is designed as an MVP but targets production‑like constraints on availability, performance, and reliability.  

| Metric                      | Symbol             | Requirement                    |
|----------------------------|--------------------|--------------------------------|
| Service availability        | \(Uptime\)         | \(Uptime \geq 99.5\%\)         |
| Recovery time objective     | \(RTO\)            | \(RTO \leq 1 \text{ h}\)       |
| Recovery point objective    | \(RPO\)            | \(RPO \leq 1 \text{ h}\)       |
| API response latency        | \(Latency_{p95}\)  | \(\leq 200 \text{ ms}\)        |
| Throughput (upload)         | \(RPS_{upload}\)   | \(\leq 100\)                   |
| Throughput (download)       | \(RPS_{download}\) | \(\leq 300\)                   |
| Concurrent users            | \(N_{active}\)     | \(\geq 1000\)                  |
| Max file size               | \(Size_{file}\)    | \(\leq 10 \text{ GB}\)         |
| Total system capacity (MVP) | \(Size_{total}\)   | \(\geq 1 \text{ TB}\)          |
| Preview processing time     | \(T_{preview}\)    | \(\leq 30 \text{ s}\)          |
| Queue capacity              | \(N_{queue}\)      | \(\leq 10^{4}\)                |
| Task success rate           | \(P_{success}\)    | \(\geq 99\%\)                  |
| Public link TTL             | \(TTL_{link}\)     | \(\leq 600 \text{ s}\)         |

These constraints inform design decisions like asynchronous task processing, separation of API and workers, and basic observability tooling.  

---

### Technology Stack

The project uses a modern Go ecosystem with standard tools for REST APIs, storage, queuing, and documentation.  

| Layer / Purpose                 | Technology / Library                               | Notes / Usage |
|---------------------------------|----------------------------------------------------|---------------|
| Programming Language            | Go (Golang)                                        | Main backend language for API and workers. |
| HTTP framework / Router         | Gin                                                | Lightweight, high‑performance HTTP router for REST. |
| API documentation              | Swag + gin‑swagger                                 | Generates and serves Swagger UI at `/swagger`. |
| Database                        | PostgreSQL                                         | Stores users, sessions, files, versions, links, events. |
| SQL Migrations                  | SQL files under `migrations/`                      | Versioned schema migrations applied on startup or via tooling. |
| ORM / Data access               | Custom repositories over `database/sql`            | Query/command repositories in `internal/infra/db`. |
| Object Storage                  | S3 / MinIO                                         | Blob storage for file contents and previews. |
| Queue / Message broker          | Kafka‑style interfaces + in‑memory mock            | Preview and event queues with mock implementation for local use. |
| Authentication / Magic links    | Custom service + SMTP email                        | Passwordless login via signed magic links and sessions. |
| Background workers              | Go binaries in `cmd/preview-worker`, `cmd/link-expirer` | Process previews, public link expiry, metrics, and events. |
| Configuration                   | YAML (`configs/`) + `.env`                         | Base and environment‑specific configuration plus overrides. |
| Logging                         | Standard logging (can be swapped to Zap/Logrus)    | Centralized logging from API and workers. |
| Metrics                         | `/api/v1/metrics` HTTP endpoint                    | Prometheus‑friendly metrics exported by a metrics worker. |
| Tracing                         | OpenTelemetry (stdout exporter)                    | Distributed tracing via middleware and SDK tracer provider. |
| Testing                         | `go test`, integration tests under `internal/test` | Unit, integration, and infrastructure tests for DB, S3, Kafka, API. |
| Containerization (optional)     | Docker / Docker Compose                            | Used to orchestrate Postgres, MinIO, and Kafka in development. |

This stack keeps the runtime lightweight while leaving room to plug in real Kafka, external SMTP, and production‑grade tracing or logging in the future.  

---

### Project Structure

The repository follows a classic Go layout with `cmd` for binaries and `internal` for private application code, reflecting hexagonal architecture and domain‑driven design principles.  

```
cmd/
  api/             # HTTP API server entrypoint
  preview-worker/  # Background preview generation worker
  link-expirer/    # Public link expiration / maintenance worker
  docs/            # Swagger docs entrypoint

internal/
  api/             # HTTP handlers, routing, middleware
  app/             # Application services, unit-of-work
  domain/          # Entities, value objects, domain events, interfaces
  infra/           # DB, queue, SMTP, storage adapters
  workers/         # Worker logic (preview, metrics, file checks, publishing)
  config/          # Configuration loading
  test/            # Integration tests (DB, S3, Kafka, API)
```

- `internal/domain` defines core entities (user, file, file version, magic link, session, public link) plus value objects and domain events.  
- `internal/app` orchestrates use cases via services, unit of work, and dependency‑free business workflows.  
- `internal/infra` provides concrete adapters for Postgres repositories, Kafka‑style queues, S3 storage, SMTP mailer, and logging.  
- `internal/api` contains Gin handlers, DTOs, presenters, and middleware for auth, error handling, and tracing.  
- `internal/workers` implements preview generation, metrics aggregation, file consistency checks, and event publishing.  

This separation keeps frameworks and drivers at the edges while protecting core domain logic from direct coupling to HTTP, SQL, or storage details.  

---

### Runtime Components

The system is a monorepo with multiple processes compiled from separate `cmd/*` packages.  

- **API Server (`cmd/api`)** — exposes REST API, serves Swagger UI, handles auth, file operations, and public link management.  
- **Preview Worker (`cmd/preview-worker`)** — consumes preview tasks from the queue, generates thumbnails, and writes them to S3/MinIO.  
- **Link Expirer (`cmd/link-expirer`)** — deactivates expired public links based on TTL and domain events.  
- **Metrics Worker** — consumes events and updates metrics exposed at `/api/v1/metrics`.  
- **File Checker Worker** — periodically verifies file storage consistency and may trigger repair or cleanup tasks.  
- **Event Publisher Worker** — pulls unprocessed domain events and publishes them to the event queue or external sinks.  

All these components share the same domain and application layers, ensuring consistent business rules across API and workers.  

---

### Architecture

The design combines layered and hexagonal patterns with domain events and queues to decouple operations and enable background processing.  

**Layered structure:**  
- **Domain layer** — entities, aggregates, value objects, and domain interfaces for storage, queues, notifications, and time.  
- **Application layer** — services implementing use cases (auth, user, file, file version, magic link, session, public link, events) plus a unit‑of‑work.  
- **Infrastructure layer** — concrete adapters for Postgres repositories, Kafka‑style queues, S3 storage, SMTP, logging, and metrics.  
- **Interface layer** — HTTP handlers, DTO mapping, presenters, middleware, and worker entrypoints.  

**Hexagonal / Ports & Adapters:**  
- Domain code depends only on small interfaces for storage, queues, notifications, and clocks.  
- The application layer coordinates domain operations and uses the unit‑of‑work to ensure consistent persistence.  
- Infrastructure adapters implement ports for Postgres, MinIO/S3, Kafka, and SMTP without leaking their APIs into domain code.  

This approach allows switching from in‑memory queues to real Kafka or from mock storage to S3/MinIO with minimal impact on business logic.  

---

### User Flow

The main user flows are upload, versioning, and sharing via short‑lived public links, with diagrams under `./docs/userflow.svg`.  

**Typical scenario:**  
1. User requests a magic link by email and receives a signed URL.  
2. User clicks the link, which is verified; a session is created and tied to device and IP information.  
3. Authenticated user uploads a file, which creates an initial `FileVersion` and stores the blob in S3/MinIO.  
4. A preview task is enqueued; the preview worker generates an image and stores it in object storage.  
5. User generates a public link with a short TTL and shares it; the link expirer worker deactivates it after expiration.  

The ER diagram at `./docs/models.svg` documents the schema and relationships between users, sessions, magic links, files, versions, events, and public links.  

---

### Security and Authentication

Authentication is fully passwordless and based on email magic links plus session management endpoints for revocation and device control.  

**Magic links:**  
- User submits email to request a login link; the system generates a signed token with an expiration time.  
- Token and metadata (user, IP, device, `expires_at`) are stored in the database and sent via SMTP.  
- When the user opens the magic link, the token is validated and a persistent session is created.  

**Sessions:**  
- Sessions are persisted with ownership, device information, IP, and expiry timestamps.  
- API exposes endpoints for listing active sessions, revoking a single session, revoking all sessions, and logging out current session.  
- Refresh token endpoint allows rotating access tokens while keeping long‑lived sessions under control.  

**Authorization and data access:**  
- Authenticated users can only access their own files and versions, enforced via foreign keys and domain checks.  
- Public link access bypasses authentication but is read‑only, time‑limited, and bound to a specific file or version.  
- Deleting a file cascades to versions, previews, public links, and related events according to schema constraints.  

Transport security is intended to be enforced via HTTPS termination at an upstream proxy or ingress in real deployments.  

---

### API Overview

The API is versioned under `/api/v1` and documented via Swagger; the Swagger UI is served at `/swagger/index.html` by the gin‑swagger middleware.  

**Public endpoints:**  

| Endpoint                         | Method | Description |
|----------------------------------|--------|-------------|
| `/swagger/*any`                  | GET    | Serve Swagger UI and OpenAPI spec. |
| `/api/v1/metrics`               | GET    | Expose metrics. |
| `/api/v1/magic-links`           | POST   | Request passwordless login link by email. |
| `/api/v1/magic-links/{token}`   | GET    | Verify magic link and create a session. |
| `/api/v1/auth/tokens/refresh`   | POST   | Refresh access token using a valid session. |
| `/api/v1/public-links/{token}`  | GET    | Download a file via public link token. |

**Authenticated endpoints (require Authorization header):**  

| Group / Endpoint                             | Method | Description |
|----------------------------------------------|--------|-------------|
| `/api/v1/auth/sessions`                      | GET    | List active sessions for current user. |
| `/api/v1/auth/sessions/{session_id}`         | DELETE | Revoke a specific session. |
| `/api/v1/auth/sessions`                      | DELETE | Revoke all sessions. |
| `/api/v1/auth/sessions/current`              | DELETE | Logout from current session. |
| `/api/v1/users/me`                           | GET    | Get current user profile and basic stats. |
| `/api/v1/users/me`                           | PATCH  | Update profile fields such as display name. |
| `/api/v1/users/me`                           | DELETE | Delete account and owned resources. |
| `/api/v1/files`                              | GET    | List user files with latest version metadata. |
| `/api/v1/files`                              | POST   | Create file and initial version, return upload URL. |
| `/api/v1/files/{file_id}`                    | GET    | Get file metadata including latest version. |
| `/api/v1/files/{file_id}`                    | PATCH  | Update file metadata (for example, rename). |
| `/api/v1/files/{file_id}`                    | DELETE | Delete file, its versions, previews, and links. |
| `/api/v1/files/{file_id}/versions`           | GET    | List versions of a file. |
| `/api/v1/files/{file_id}/versions`           | POST   | Upload a new version, return upload URL. |
| `/api/v1/files/{file_id}/versions/{num}/content` | GET | Get signed download URL for a specific version. |
| `/api/v1/files/{file_id}/versions/{num}/restore` | POST | Mark a specific version as the latest. |
| `/api/v1/files/{file_id}/public-links`       | GET    | List active public links for a file. |
| `/api/v1/files/{file_id}/public-links`       | POST   | Create a new public link with TTL. |
| `/api/v1/files/{file_id}/public-links/{link_id}` | DELETE | Revoke a specific public link. |

The exact request and response schemas, including DTOs and error formats, are defined in the generated Swagger spec under `./docs/swagger.yaml` and `./docs/swagger.json`.  

---

### Observability (Metrics and Tracing)

Observability is built into the service through a metrics endpoint and OpenTelemetry tracing configured in the API server.  

- The API exposes `/api/v1/metrics`, which is intended to be scraped by Prometheus or compatible systems.  
- A metrics worker consumes events and aggregates counters or histograms for requests, tasks, and domain operations.  
- OpenTelemetry is initialized in `cmd/api/main.go` using a stdout exporter and tracer provider, and a tracing middleware wraps the request lifecycle.  
- Traces can later be redirected to Jaeger, Tempo, or another backend by swapping the exporter.  

This setup enables basic monitoring of latency, throughput, and failure rates without changing the business logic.  

---

### Local Development

Local development assumes a running Postgres instance plus optional MinIO and Kafka‑like services, which can be orchestrated via Docker Compose.  

Typical workflow:  
1. Copy and adjust configuration files under `configs/config.base.yaml` and `configs/config.dev.yaml`, plus `.env` for secrets.  
2. Start dependencies (Postgres, MinIO, queue) locally or via Docker.  
3. Run migrations from the `migrations/` directory using your preferred migration tool.  
4. Start the API server using `go run ./cmd/api`.  
5. Optionally start `go run ./cmd/preview-worker` and `go run ./cmd/link-expirer` for background processing.  
6. Open `/swagger/index.html` to explore and test the API interactively.  

Tests can be executed via `go test ./...`, including integration tests that interact with Postgres, MinIO, and the queue mocks.  

---

### Future Improvements

While the current design already separates core logic, infrastructure, and runtime components, several improvements are natural next steps.  

- Replace in‑memory queues with a real Kafka cluster in non‑test environments.  
- Add antivirus scanning, content classification, and webhooks as additional asynchronous tasks.  
- Introduce a CDN in front of file downloads and previews for global performance.  
- Enhance search with full‑text indexing and tags for large file collections.  
- Harden security with optional 2FA, stricter token policies, and per‑device approvals.  
- Wire tracing and metrics into a full observability stack using Prometheus, Loki, and Tempo or Jaeger.  

The existing hexagonal structure and clear separation of API, workers, and infrastructure make it straightforward to evolve this MVP toward a more distributed or microservice‑oriented architecture if future load requires it.  
```