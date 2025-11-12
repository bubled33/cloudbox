# Cloud File Storage (Google Drive Lite)

> **Language / Язык**: **EN** | [RU](./docs/README_RU.md)

---

## Introduction

**Project**: Cloud File Storage (Google Drive Lite)

**Objective**: Development of a web service for storing, managing, and sharing files with support for versioning and preview generation.

| Metric                      | Symbol           | Requirement            |
| --------------------------- | ---------------- | ---------------------- |
| Service availability        | $Uptime$         | $Uptime \geq 99.5\%$   |
| Recovery time objective     | $RTO$            | $RTO \leq 1 \text{ h}$ |
| Recovery point objective    | $RPO$            | $RPO \leq 1 \text{ h}$ |
| API response latency        | $Latency_{p95}$  | $\leq 200 \text{ ms}$  |
| Throughput (upload)         | $RPS_{upload}$   | $\leq 100$             |
| Throughput (download)       | $RPS_{download}$ | $\leq 300$             |
| Concurrent users            | $N_{active}$     | $\geq 1000$            |
| Max file size               | $Size_{file}$    | $\leq 10 \text{ GB}$   |
| Total system capacity (MVP) | $Size_{total}$   | $\geq 1 \text{ TB}$    |
| Preview processing time     | $T_{preview}$    | $\leq 30 \text{ s}$    |
| Queue capacity              | $N_{queue}$      | $\leq 10^{4}$          |
| Task success rate           | $P_{success}$    | $\geq 99\%$            |
| Public link TTL             | $TTL_{link}$     | $\leq 600 \text{ s}$   |

<p>&nbsp;</p>

## Technology Stack

The Cloud File Storage system is built using a **modern Go ecosystem**, combining high-performance web frameworks, reliable databases, scalable queues, and background workers.

| Layer / Purpose               | Technology / Library                          | Notes / Usage |
|-------------------------------|-----------------------------------------------|---------------|
| **Programming Language**      | Go (Golang)                                   | Main backend language, compiles to a single binary, suitable for high-performance services |
| **Web Framework / Router**    | [Gin](https://github.com/gin-gonic/gin)       | High-performance HTTP framework, lightweight and convenient for REST APIs |
| **Database**                  | PostgreSQL                                    | Relational database for storing users, files, file versions, and links |
| **ORM / Database Access**     | [GORM](https://gorm.io/)                       | Popular Go ORM, supports migrations and table relationships |
| **Object Storage**            | S3 / MinIO                                    | File storage (up to 10 GB), MinIO for local development, S3 for production |
| **Queue / Message Broker**    | Apache Kafka                                  | Asynchronous task processing (Preview Worker, Public Link Expirer) |
| **Authentication / Magic Link** | JWT / custom magic link flow                 | Passwordless login via temporary links, tokens stored in the database |
| **Background Workers**        | Go routines + Kafka consumers                 | Asynchronous task processing, scalable with multiple worker instances |
| **Configuration / Env**       | Viper / Envconfig                             | Configuration management, environment variables |
| **Logging / Monitoring**      | Logrus / Zap                                  | Logging for debugging and monitoring |
| **Testing / QA**              | Go test, Testify                               | Unit tests and integration tests |
| **Containerization / Deployment** | Docker / Docker Compose                     | Simplified local development and production deployment |
| **CI/CD**                     | GitHub Actions / GitLab CI                     | Automated build, tests, and deployment |

> 💡 **Notes:**  
> - **Gin + GORM** is a lightweight and standard stack for Go web services.  
> - **Kafka** allows scaling background workers and handling high task volumes asynchronously.  
> - **MinIO** is used for local development instead of S3.  
> - **Workers** use Go routines and Kafka consumers to keep API responsive while processing heavy tasks asynchronously.

<p>&nbsp;</p>

## Architecture

**Project Architecture**: Cloud File Storage is implemented as a **monolithic application** using a layered and hexagonal approach.  
This structure ensures clear separation of responsibilities, testability, and maintainability while keeping deployment simple.

**Layered Structure**:
- **Presentation / API Layer** – handles HTTP requests, magic links, file upload/download, and public link generation.
- **Application / Service Layer** – coordinates file operations, versioning, link management, and background tasks.
- **Domain / Core Layer** – contains business logic and entities (User, File, FileVersion, MagicLink, PublicLink).
- **Infrastructure Layer** – concrete implementations of external dependencies: database (Postgres), object storage (S3/MinIO), task queue (Redis/RabbitMQ/Kafka).

**Hexagonal (Ports & Adapters) Principles**:
- The **Domain Layer** is independent of external services.
- **Ports** define interfaces for database access, storage operations, and queue interactions.
- **Adapters** implement these interfaces, allowing easy replacement of infrastructure components without changing business logic.

**Background Workers**:
- **Preview Worker** – asynchronously generates file previews by consuming tasks from the queue.
- **Public Link Expirer** – deactivates public links after TTL expires.
- Workers run inside the monolith but interact through the task queue for asynchronous execution and scalability.

> ⚠️ **Monolith vs Microservices:**  
> Although modern cloud systems often use **microservices**, this project is implemented as a **monolith** for practical reasons.  
> 
> - **Simplicity** – keeping all business logic, API routes, and background workers in a single application makes the code easier to understand and maintain.  
> - **Ease of Deployment** – only one service needs to be deployed, while external dependencies like the database, object storage, and task queue remain separate.  
> - **Testing and Development** – running a monolith locally is straightforward, without managing multiple interdependent services.  
> - **Scalability of Critical Components** – even in a monolithic setup, the **Preview Worker** and **Public Link Expirer** can scale independently via multiple processes and queues.  
> - **Future Flexibility** – the codebase is organized with **Layered + Hexagonal architecture**, so it can be refactored into microservices if needed.

<p>&nbsp;</p>

![Component Diagram](./docs/architecture.svg)

<p>&nbsp;</p>

## User Flow

The following diagram illustrates the key user scenario in **Cloud File Storage**: uploading a file and generating a temporary public link.  

<p>&nbsp;</p>

![Component Diagram](./docs/userflow.svg)

<p>&nbsp;</p>

**Explanation of Components:**

- **User / App** – initiates actions such as login and file upload.  
- **DB: Files / FileVersions** – stores metadata and version history.  
- **S3 / Object Storage** – stores actual file content and generated previews.  
- **Queue: Preview Tasks** – manages asynchronous preview generation.  
- **Preview Worker** – consumes tasks from the queue to generate previews.  
- **Generate Public Link** – creates temporary links for file sharing.  

> ⚠️ **Note:** Only the **upload file** and **generate public link** endpoints are shown in this flow.  
> The full system contains additional endpoints (versioning, download, deletion, etc.), but they are omitted here for clarity, focusing on the main user interaction path.

<p>&nbsp;</p>

## ER Diagram

The diagram below shows the main entities of the **Cloud File Storage** system (Google Drive Lite), including users, files, file versions, sessions, magic links, and public links.  
It illustrates the relationships between tables and the cardinality of each association.

**Legend:**
- `{PK}` = Primary Key  
- `{FK}` = Foreign Key  
- `ON DELETE CASCADE` = dependent rows are automatically deleted  
- `1` = one, `*` = many  
- Denormalization: some fields (e.g., `name`, `size`, `mime` in `files`) are duplicated from `file_versions` for faster access.

> ⚠️ **Note on normalization:**  
> The `files` table duplicates some attributes from `file_versions` (name, size, mime, etc.).  
> This technically violates **3rd Normal Form (3NF)** because these attributes depend on the related `file_versions` record, not directly on the `files` primary key.  
> This denormalization is intentional to speed up queries for listing files, showing file metadata, and avoiding frequent joins. In production systems, this trade-off is common for performance reasons.

<p>&nbsp;</p>

![ER Diagram](./docs/models.svg)

<p>&nbsp;</p>

## Security / Authentication

The Cloud File Storage system implements **passwordless authentication** and several security measures to protect user data, files, and public links.

### Authentication

- **Magic Links (Passwordless Login)**  
  - Users authenticate via temporary magic links sent to their email.  
  - Each magic link contains a **signed token** with an expiration time (TTL).  
  - Once verified, a **session token (JWT)** is issued for API requests.  
  - No passwords are stored in the database, reducing the risk of credential leaks.

- **Session Management**  
  - Sessions are stored in the database with expiration times.  
  - API requests must include a valid session token in the `Authorization` header.  
  - Expired sessions require users to request a new magic link.

### Authorization

- **File Access Control**  
  - Users can only access files they own or files shared via valid public links.  
  - Database foreign key constraints and application logic enforce ownership rules.

- **Public Links**  
  - Temporary links have a **TTL ≤ 600 seconds**.  
  - **Public Link Expirer** automatically deactivates expired links.  
  - Access via public links bypasses authentication but is limited to read-only operations.

### Data Protection

- **Transport Security**  
  - All API traffic must use HTTPS.  
- **File Storage Security**  
  - S3/MinIO buckets can be configured with server-side encryption.  
- **Sensitive Data Handling**  
  - Only minimal metadata is stored in the database; no plaintext passwords.  

> 💡 **Notes:**  
> - The combination of passwordless login, TTL-based links, and secure transport ensures both usability and security.  
> - Additional measures, such as 2FA or encryption at rest for files, can be added in future improvements.

<p>&nbsp;</p>

## API Documentation

The Cloud File Storage system exposes a **RESTful API** to manage users, files, file versions, magic links, and public links. All endpoints use **JSON** for request and response payloads.

| Endpoint                                   | Method | Description                                         | Request Body / Params | Response |
|--------------------------------------------|--------|-----------------------------------------------------|---------------------|---------|
| `/auth/magic-link`                          | POST   | Request a passwordless login link for a user       | `{ "email": "user@example.com" }` | `{ "magic_link": "https://..." }` |
| `/auth/magic-link/verify`                   | POST   | Verify magic link and create a session             | `{ "token": "..." }` | `{ "session_id": "..." }` |
| `/files/upload`                             | POST   | Upload a **new file** and create its initial version | `{ "name": "file.pdf", "size": 1024 }` | `{ "file_id": 123, "upload_url": "https://..." }` |
| `/files/{file_id}/upload`                   | POST   | Upload a **new version** of an existing file       | Path param: `file_id`, `{ "size": 2048 }` | `{ "version_id": 2, "upload_url": "https://..." }` |
| `/files/{file_id}`                           | GET    | Retrieve file metadata (latest version, owner, name, size, created_at) | Path param: `file_id` | `{ "id": 123, "name": "...", "size": 1024, "owner_id": 1, "latest_version": 2, "created_at": "..." }` |
| `/files/{file_id}/rename`                   | PATCH  | Rename a file                                      | Path param: `file_id`, `{ "name": "new_name.pdf" }` | `{ "status": "renamed", "new_name": "new_name.pdf" }` |
| `/files/download/{file_id}`                 | GET    | Download the **latest version** of a file          | Path param: `file_id` | File stream |
| `/files/{file_id}/versions`                 | GET    | List all versions of a file                         | Path param: `file_id` | `[ { "version_id": 1, "size": 1024, "created_at": "..." }, ... ]` |
| `/files/{file_id}/versions/{version_id}`    | GET    | Retrieve metadata of a specific version           | Path param: `file_id`, `version_id` | `{ "version_id": 2, "size": 2048, "created_at": "..." }` |
| `/files/{file_id}/versions/{version_id}/restore` | POST | Restore a specific version as the latest version   | Path param: `file_id`, `version_id` | `{ "status": "restored" }` |
| `/files/{file_id}/versions/{version_id}`    | DELETE | Delete a specific version of a file               | Path param: `file_id`, `version_id` | `{ "status": "deleted" }` |
| `/files/{file_id}/preview`                  | GET    | Get the **preview** for the latest version of a file | Path param: `file_id` | Image/Thumbnail stream |
| `/files/{file_id}`                           | DELETE | Delete a file and **all its versions**, including previews and public links | Path param: `file_id` | `{ "status": "deleted" }` |
| `/public-links`                             | POST   | Generate a temporary public link for a file       | `{ "file_id": 123, "ttl": 600 }` | `{ "public_link": "https://..." }` |
| `/public-links/{link_id}`                   | GET    | Access a file via public link                      | Path param: `link_id` | File stream |
| `/public-links/{link_id}`                   | DELETE | Delete a specific public link                      | Path param: `link_id` | `{ "status": "deleted" }` |
| `/users/me`                                 | GET    | Retrieve current user profile                      | Auth token in header | `{ "id": 1, "email": "user@example.com", "files": [...] }` |

> 💡 **Notes:**  
> - **File Versioning:** Each upload to `/files/{file_id}/upload` creates a new `FileVersion` record, while the main `Files` table keeps metadata for the latest version.  
> - **Deletion:**  
>   - `/files/{file_id}` removes the main file, all versions, previews, and associated public links (`ON DELETE CASCADE`).  
>   - `/files/{file_id}/versions/{version_id}` deletes a single version.  
>   - `/public-links/{link_id}` allows manual removal of a single public link.  
> - **Restore Version:** `/files/{file_id}/versions/{version_id}/restore` sets a previous version as the latest version.  
> - **Preview Generation:** Previews are generated asynchronously; they may be temporarily unavailable after a new version is uploaded.  
> - **Public Links:** Respect `TTL ≤ 600 seconds` and expire automatically via the **Public Link Expirer**.  
> - All endpoints except public links require authentication via session token.

<p>&nbsp;</p>

## Future Improvements / Scalability

The Cloud File Storage system is designed as an MVP with a **monolithic architecture**, but several improvements can be implemented in the future to enhance functionality, performance, and scalability.

| Area                        | Possible Improvements / Enhancements                 | Notes |
|------------------------------|------------------------------------------------------|-------|
| **Asynchronous Tasks**       | Add antivirus scanning, notifications (email/webhook), temporary file cleanup | Offload additional heavy tasks to background workers via the queue to keep API responsive |
| **File Storage / CDN**       | Integrate a CDN for file delivery and preview assets | Improves download speed for end users globally |
| **Database Scaling**         | Implement read replicas, partitioning, or sharding | Allows handling higher user and file volume while maintaining low latency |
| **Microservices Transition** | Split monolith into microservices (e.g., API, preview generation, link management) | Provides independent deployment, scaling, and fault isolation |
| **Advanced Search & Indexing** | Full-text search for file metadata, tags, and content indexing | Improves user experience for large datasets |
| **Security Enhancements**    | Add encryption at rest for files, 2FA for users, improved token management | Increases protection for sensitive data |
| **Monitoring / Observability** | Implement Prometheus + Grafana for metrics, alerts, and dashboards | Helps detect bottlenecks and maintain uptime ≥ 99.5% |
| **Horizontal Scaling of Workers** | Deploy multiple Preview Worker and Expirer instances | Ensures processing capacity meets peak demand (up to 10k tasks in queue) |

> 💡 **Notes:**  
> - These improvements are optional for MVP but planned for future versions.  
> - The current **Layered + Hexagonal architecture** allows easy refactoring to microservices or scaling individual components without major rewrites.  
> - Priority should be given to asynchronous task scaling and storage optimization to maintain performance and responsiveness for 1000+ concurrent users.

