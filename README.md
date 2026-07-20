# Loglift

A self-built mini-Datadog: structured log ingestion, buffering, indexing,
dashboarding, and threshold alerting — built to understand how real
observability platforms (Datadog, ELK, Grafana) work under the hood.

## Stack
- Go — log agent, indexer, alert worker
- Redis Streams — ingestion buffer
- OpenSearch — log storage and search
- React — dashboard
- Docker Compose (dev) / Kubernetes (deploy)



## Local dev
\`\`\`bash
docker compose up -d
\`\`\`