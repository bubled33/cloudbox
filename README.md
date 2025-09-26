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

![Component Diagram](./architecture.svg)

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

![ER Diagram](./models.svg)
