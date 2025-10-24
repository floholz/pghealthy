# pghealthy 🩺🐘

A tiny HTTP service that checks the health of a PostgreSQL instance. Perfect for containers and Kubernetes. It exposes a simple /healthz endpoint that verifies:

- Plug-and-play connectivity to PostgreSQL via DSN or individual env vars 🔌
- Optional table existence checks 📋
- Optional custom SQL queries, with opt-in result exposure 🧪

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
![Go Version](https://img.shields.io/badge/Go-%E2%89%A51.20-00ADD8?logo=go)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)

---

- Jump to: [Quick start](#-quick-start) • [API](#-api) • [Configuration](#-configuration) • [Examples](#-examples) • [Docker](#-docker) • [Docker Compose](#-docker-compose) • [Build](#-build-from-source) • [Operational notes](#-operational-notes) • [License](#-license)

## 🚀 Quick start

- Run locally (requires Go installed):
  ```bash
  go run .
  # Defaults: listens on :2345 and connects to \
  # postgres://postgres@localhost:5432/postgres?sslmode=disable
  ```

- Curl the health endpoint:
  ```bash
  curl -i http://localhost:2345/healthz
  ```

You should see HTTP/1.1 200 OK and a JSON body: {"status":"ok"}

## 🔎 Response format

- 200 OK: {"status":"ok", "results":[...]} — results included only when PG_HEALTHY_EXPOSE_QUERY_RESULTS is true and queries are configured
- 503 Service Unavailable: {"status":"unhealthy","error":"..."}

## 🔧 Configuration

You can configure the service with either a full PostgreSQL connection string or component env vars.

### Connection settings

- PG_CONNECTION_STRING
  - Example: postgres://user:pass@host:5432/dbname?sslmode=disable
- OR the following individual variables (used to build the DSN when PG_CONNECTION_STRING is not provided):
  - POSTGRES_USER (default: postgres)
  - POSTGRES_PASSWORD (no default)
  - POSTGRES_DB (default: postgres)
  - POSTGRES_HOST (default: localhost)
  - POSTGRES_PORT (default: 5432)
  - POSTGRES_SSLMODE (default: disable)

### Health behavior

- PG_HEALTHY_TABLES
  - Comma-separated list of table names to verify exist in information_schema.tables
  - Example: PG_HEALTHY_TABLES=User,Session,Company
- PG_HEALTHY_QUERIES
  - Custom SQL queries to run; use ';;' to separate multiple queries
  - Each query must return exactly one row and one column (a single scalar value)
  - Example: PG_HEALTHY_QUERIES=SELECT "ID" FROM "User" LIMIT 1;;SELECT 1
- PG_HEALTHY_EXPOSE_QUERY_RESULTS
  - When true or 1, includes the query scalar results array in the response JSON
  - Default: false
- PORT
  - HTTP listen port (default: 2345)

## 📚 Examples

- Minimal with DSN:
  ```bash
  PG_CONNECTION_STRING=postgres://postgres:secret@db:5432/app?sslmode=disable pghealthy
  ```
- Component vars and table checks:
  ```bash
  POSTGRES_USER=postgres POSTGRES_PASSWORD=secret POSTGRES_DB=app POSTGRES_HOST=db \
    PG_HEALTHY_TABLES=users,orders pghealthy
  ```
- With queries and exposed results:
  ```bash
  PG_CONNECTION_STRING=postgres://postgres:secret@db:5432/app?sslmode=disable \
    PG_HEALTHY_QUERIES='SELECT COUNT(*) FROM users;;SELECT 1' \
    PG_HEALTHY_EXPOSE_QUERY_RESULTS=1 pghealthy
  ```

## 🐳 Docker

A minimal container image is defined by Dockerfile.

- Build locally:
  ```bash
  docker build -t pghealthy:local .
  ```
- Run:
  ```bash
  docker run --rm -p 2345:2345 \
    -e PG_CONNECTION_STRING='postgres://postgres:secret@host.docker.internal:5432/postgres?sslmode=disable' \
    pghealthy:local
  ```

## 🧩 Docker Compose

A minimal example compose file is included; adapt it to fit your environment and needs.

- docker-compose.yml (local example):
  ```yaml
  services:
    pghealthy:
      build: .
      environment:
        - POSTGRES_USER=postgres
        - POSTGRES_PASSWORD=my_local_password
        - POSTGRES_DB=my_database
        - POSTGRES_HOST=postgres_database
        - POSTGRES_PORT=5432
        - PG_HEALTHY_TABLES=User,Documents
      restart: unless-stopped
    database:
      image: postgres:14.1-alpine
      environment:
        - POSTGRES_PASSWORD=my-database-password
      volumes:
        - ./data:/var/lib/postgresql/data
  ```

## 🏗️ Build from source

- Requirements: Go 1.20+ (module file provided; container uses golang:1.25-alpine)
- Steps:
  ```bash
  go build -o pghealthy .
  ./pghealthy
  ```

## 🧭 Operational notes

- Connection pooling: low footprint defaults (MaxOpenConns=2, MaxIdleConns=1, ConnMaxLifetime=2m)
- Logs: service logs to stdout; includes DSN (without password if not present in components) and health outcomes
- Security: Do not enable PG_HEALTHY_EXPOSE_QUERY_RESULTS unless you are comfortable returning query outputs via your health endpoint. Configure network access accordingly.

## 🔗 API

- GET /healthz
  - 200 OK when all checks pass
  - 503 Service Unavailable when Ping, table checks, or queries fail

## 📄 License

- MIT License. See LICENSE for details.

## 🙌 Credits

- Built by floholz. Container image labels reference https://github.com/floholz/pghealthy
