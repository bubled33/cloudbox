Project: Cloud File Storage (Google Drive Lite)
Objective: Development of a web service for storing, managing, and sharing files with support for versioning and preview generation.

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
