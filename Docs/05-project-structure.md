# Employee Management Platform — Project Structure

## 1. Target Repository

Suggested repository name:

```text
employee-management-platform
```

The repository is completely independent of the existing `go-task-api` project.

## 2. Proposed Structure

```text
employee-management-platform/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── database/
│   │   └── mysql.go
│   │
│   ├── model/
│   │   └── employee.go
│   │
│   ├── repository/
│   │   └── employee_repository.go
│   │
│   ├── service/
│   │   └── employee_service.go
│   │
│   ├── handler/
│   │   └── employee_handler.go
│   │
│   ├── middleware/
│   │   └── ...
│   │
│   └── router/
│       └── router.go
│
├── migrations/
│
├── tests/
│
├── deployments/
│   └── kubernetes/
│
├── helm/
│
├── docs/
│
├── Dockerfile
├── docker-compose.yml
├── Jenkinsfile
├── Makefile
├── .dockerignore
├── .gitignore
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

## 3. Package Responsibilities

### `cmd/server`

Application entry point only.

It should assemble configuration, database, repositories, services, handlers, and router.

### `internal/config`

Loads and validates runtime configuration.

### `internal/database`

Creates and manages the MySQL connection/pool.

### `internal/model`

Contains domain/database models and related types.

### `internal/repository`

Persistence abstraction.

### `internal/service`

Business logic.

### `internal/handler`

HTTP layer.

### `internal/router`

Gin routes and middleware registration.

### `migrations`

Version-controlled database schema changes.

### `deployments`

Kubernetes manifests if raw manifests are used.

### `helm`

Helm chart for repeatable Kubernetes deployment.

The exact timing of Helm introduction will be decided after basic Kubernetes deployment works.

## 4. Dependency Direction

Preferred direction:

```text
handler
   ↓
service
   ↓
repository
   ↓
database
```

Lower-level packages should not depend on HTTP handlers.

## 5. Entry Point Principle

`main.go` should remain small.

It should primarily:

1. Load configuration.
2. Initialize database.
3. Initialize repository.
4. Initialize service.
5. Initialize handlers/router.
6. Start HTTP server.
7. Handle graceful shutdown.
