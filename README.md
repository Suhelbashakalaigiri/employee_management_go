# Employee Management Platform — Version 1 (V1)

A clean, production-oriented Go REST API for Employee Management built using idiomatic Go, Gin, and MySQL. This repository represents **Version 1 (Application Core)** of the multi-phase deployment journey:

```text
Go + Gin (V1)
   ↓
MySQL (V1)
   ↓
Docker
   ↓
Git / GitHub
   ↓
Jenkins CI/CD
   ↓
Kubernetes
   ↓
AWS ECR + EKS + RDS MySQL
```

---

## 1. Project Purpose

The Employee Management Platform is a production-oriented learning project whose business capability is intentionally focused: managing employees through a clean, standard RESTful API.

The primary objective is to learn how a real-world Go application progresses from local code to containerization, automated testing, continuous integration, Kubernetes orchestration, and cloud deployment on AWS EKS and RDS.

---

## 2. Technology Stack

* **Language:** [Go](https://golang.org/) (Go 1.27+)
* **Web Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
* **Database:** MySQL 8.0+
* **Driver:** `github.com/go-sql-driver/mysql` (Standard `database/sql` driver with connection pooling)
* **Fixed-Point Decimal:** `github.com/shopspring/decimal` (Precise financial/monetary storage for salaries without float precision loss)
* **Testing:** Go standard library `testing`, `net/http/httptest`, and `github.com/DATA-DOG/go-sqlmock`

---

## 3. Architecture Overview

The application strictly implements a decoupled, 4-tier layered architecture:

```text
Client
  │ (HTTP / JSON)
  ▼
Gin Router & Middleware (Logger, Panic Recovery, Route Not Found)
  │
  ▼
Handler Layer (`internal/handler`)
  │ - Parses HTTP parameters and validates payloads
  │ - Translates domain outcomes to HTTP status codes & JSON contracts
  ▼
Service Layer (`internal/service`)
  │ - Coordinates business logic, validation, and email uniqueness checks
  │ - Completely decoupled from Gin HTTP contexts
  ▼
Repository Layer (`internal/repository`)
  │ - Parameterized SQL queries against MySQL
  │ - Encapsulates database-specific behavior (e.g., MySQL error 1062 translation)
  ▼
MySQL Database (`internal/database`)
    - Configured connection pool (open conns, idle conns, timeouts)
```

### Key Architectural Characteristics
* **Centralized Configuration:** All configuration is driven by environment variables with sensible defaults and strict startup validation.
* **Consistent Error Contract:** All errors follow a uniform JSON schema (`{"error": {"code": "...", "message": "...", "details": {...}}}`).
* **Information Security:** Raw database errors, passwords, queries, and stack traces are masked from client responses.
* **Graceful Shutdown:** Intercepts `SIGINT` and `SIGTERM`, shuts down the HTTP server within a 10-second timeout context, drains active requests, closes database connections cleanly, and exits.

---

## 4. Project Structure

```text
employee-management-go/
├── cmd/
│   └── server/
│       └── main.go                         # Application entry point & graceful shutdown
├── internal/
│   ├── config/
│   │   ├── config.go                       # Environment configuration loader & validator
│   │   └── config_test.go                  # Configuration unit tests
│   ├── database/
│   │   └── mysql.go                        # MySQL connection pooling & verification
│   ├── errors/
│   │   └── errors.go                       # Centralized application error types & codes
│   ├── handler/
│   │   ├── employee_handler.go             # Employee REST handlers
│   │   ├── employee_handler_test.go        # Handler & API integration tests
│   │   ├── health_handler.go               # /health endpoint handler
│   │   └── health_handler_test.go          # Health endpoint tests
│   ├── middleware/
│   │   ├── logger.go                       # HTTP request latency/status logger
│   │   └── recovery.go                     # Panic recovery with sanitized error responses
│   ├── model/
│   │   ├── date.go                         # Custom Date type (YYYY-MM-DD) for JSON & SQL
│   │   ├── employee.go                     # Employee entity, DTOs & validation logic
│   │   └── employee_test.go                # Model validation & Date tests
│   ├── repository/
│   │   ├── employee_repository.go          # MySQL persistence implementation
│   │   └── employee_repository_test.go     # Deterministic SQL repository tests with sqlmock
│   ├── router/
│   │   └── router.go                       # Gin route registrations & middleware assembly
│   └── service/
│       ├── employee_service.go             # Employee business logic & rules
│       └── employee_service_test.go        # Service layer isolated unit tests
├── migrations/
│   ├── 000001_create_employees.up.sql      # Schema creation migration
│   └── 000001_create_employees.down.sql    # Schema rollback migration
├── .env.example                            # Example environment configuration
├── .gitignore                              # Git exclusion rules
├── go.mod                                  # Go module definitions
├── go.sum                                  # Dependency checksums
└── README.md                               # Project documentation
```

---

## 5. Prerequisites

* **Go:** Version 1.24+ (Tested on Go 1.27)
* **MySQL:** Version 8.0+

---

## 6. Environment Variables

Configure the application by creating a `.env` file or exporting environment variables:

| Variable | Type | Default | Description |
|---|---|---|---|
| `APP_PORT` | int | `8080` | HTTP port for server binding |
| `APP_ENV` | string | `development` | Runtime environment (`development`, `production`) |
| `DB_HOST` | string | `localhost` | MySQL host address |
| `DB_PORT` | int | `3306` | MySQL port |
| `DB_NAME` | string | `employee_management` | MySQL database name |
| `DB_USER` | string | `root` | Database username |
| `DB_PASSWORD` | string | `""` | Database password |
| `DB_MAX_OPEN_CONNS` | int | `25` | Maximum number of open connections |
| `DB_MAX_IDLE_CONNS` | int | `10` | Maximum number of idle connections |
| `DB_CONN_MAX_LIFETIME_MINUTES` | int | `5` | Maximum connection lifetime in minutes |
| `DB_CONN_MAX_IDLE_TIME_MINUTES` | int | `5` | Maximum connection idle time in minutes |

A template is provided in [.env.example](.env.example).

---

## 7. MySQL Setup

1. Start your local MySQL server.
2. Log in and create the database:

```sql
CREATE DATABASE IF NOT EXISTS employee_management CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

3. (Optional) Create a dedicated user:

```sql
CREATE USER IF NOT EXISTS 'employee_user'@'localhost' IDENTIFIED BY 'secretpassword';
GRANT ALL PRIVILEGES ON employee_management.* TO 'employee_user'@'localhost';
FLUSH PRIVILEGES;
```

---

## 8. Database Migration Instructions

The schema changes are version-controlled inside the [migrations/](migrations/) folder.

To apply the migration manually via the MySQL CLI:

```bash
# On Linux / macOS / Git Bash:
mysql -u root -p employee_management < migrations/000001_create_employees.up.sql

# On Windows PowerShell:
Get-Content migrations/000001_create_employees.up.sql | mysql -u root -p employee_management
```

Or using the standard `golang-migrate` CLI:

```bash
migrate -path migrations -database "mysql://root:password@tcp(localhost:3306)/employee_management" up
```

To rollback:

```bash
mysql -u root -p employee_management < migrations/000001_create_employees.down.sql
```

---

## 9. How to Run the Application

Once your MySQL instance is running and migrations have been executed:

```bash
# Set your environment variables (or use defaults)
export DB_PASSWORD="your_password"

# Run the server
go run ./cmd/server
```

On Windows PowerShell:

```powershell
$env:DB_PASSWORD="your_password"
go run ./cmd/server
```

---

## 10. How to Run Tests

Run the full automated test suite across all layers:

```bash
# Run all unit and integration tests
go test -v ./...

# Run code vetting
go vet ./...

# Run formatting check
gofmt -s -l .
```

All repository tests use deterministic in-memory SQL mocking (`go-sqlmock`), so tests execute instantly without requiring a live MySQL server.

---

## 11. API Endpoint Summary

| Method | Endpoint | Description | Status Code |
|---|---|---|---|
| `GET` | `/health` | Lightweight application liveness probe | `200 OK` |
| `POST` | `/api/v1/employees` | Create a new employee | `201 Created` |
| `GET` | `/api/v1/employees` | List all employees | `200 OK` |
| `GET` | `/api/v1/employees/:id` | Get employee details by ID | `200 OK` / `404 Not Found` |
| `PUT` | `/api/v1/employees/:id` | Update mutable fields of an employee | `200 OK` / `404 Not Found` |
| `DELETE` | `/api/v1/employees/:id` | Delete an employee | `204 No Content` / `404 Not Found`|

---

## 12. Example API Requests

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health
```
**Response (200 OK):**
```json
{
  "status": "UP"
}
```

### 2. Create Employee
```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Rahul",
    "last_name": "Kumar",
    "email": "rahul.kumar@example.com",
    "phone": "+919876543210",
    "department": "Engineering",
    "designation": "Software Engineer",
    "salary": 65000,
    "joining_date": "2026-09-24"
  }'
```
**Response (201 Created):**
```json
{
  "id": 1,
  "first_name": "Rahul",
  "last_name": "Kumar",
  "email": "rahul.kumar@example.com",
  "phone": "+919876543210",
  "department": "Engineering",
  "designation": "Software Engineer",
  "salary": "65000",
  "joining_date": "2026-09-24",
  "created_at": "2026-09-28T10:00:00Z",
  "updated_at": "2026-09-28T10:00:00Z"
}
```

### 3. Validation Error Example
```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "",
    "email": "invalid-email"
  }'
```
**Response (400 Bad Request):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid employee data",
    "details": {
      "first_name": "first_name is required",
      "email": "must be a valid email address",
      "last_name": "last_name is required",
      "phone": "phone is required",
      "department": "department is required",
      "designation": "designation is required",
      "joining_date": "joining_date is required and must be in YYYY-MM-DD format"
    }
  }
}
```

### 4. Duplicate Email Conflict Example
```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Rahul",
    "last_name": "Kumar",
    "email": "rahul.kumar@example.com",
    "phone": "+919876543210",
    "department": "Engineering",
    "designation": "Software Engineer",
    "salary": 65000,
    "joining_date": "2026-09-24"
  }'
```
**Response (409 Conflict):**
```json
{
  "error": {
    "code": "EMPLOYEE_EMAIL_CONFLICT",
    "message": "An employee with this email already exists"
  }
}
```

### 5. List Employees
```bash
curl -X GET http://localhost:8080/api/v1/employees
```
**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "first_name": "Rahul",
      "last_name": "Kumar",
      "email": "rahul.kumar@example.com",
      "phone": "+919876543210",
      "department": "Engineering",
      "designation": "Software Engineer",
      "salary": "65000",
      "joining_date": "2026-09-24",
      "created_at": "2026-09-28T10:00:00Z",
      "updated_at": "2026-09-28T10:00:00Z"
    }
  ]
}
```

### 6. Get Employee by ID
```bash
curl -X GET http://localhost:8080/api/v1/employees/1
```
**Response (200 OK):**
```json
{
  "id": 1,
  "first_name": "Rahul",
  "last_name": "Kumar",
  "email": "rahul.kumar@example.com",
  "phone": "+919876543210",
  "department": "Engineering",
  "designation": "Software Engineer",
  "salary": "65000",
  "joining_date": "2026-09-24",
  "created_at": "2026-09-28T10:00:00Z",
  "updated_at": "2026-09-28T10:00:00Z"
}
```

### 7. Update Employee
```bash
curl -X PUT http://localhost:8080/api/v1/employees/1 \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Rahul",
    "last_name": "Verma",
    "email": "rahul.verma@example.com",
    "phone": "+919876543210",
    "department": "Engineering",
    "designation": "Staff Software Engineer",
    "salary": 95000,
    "joining_date": "2026-09-24"
  }'
```
**Response (200 OK):**
```json
{
  "id": 1,
  "first_name": "Rahul",
  "last_name": "Verma",
  "email": "rahul.verma@example.com",
  "phone": "+919876543210",
  "department": "Engineering",
  "designation": "Staff Software Engineer",
  "salary": "95000",
  "joining_date": "2026-09-24",
  "created_at": "2026-09-28T10:00:00Z",
  "updated_at": "2026-09-28T10:05:00Z"
}
```

### 8. Delete Employee
```bash
curl -X DELETE http://localhost:8080/api/v1/employees/1
```
**Response:** `204 No Content`

### 9. Employee Not Found Example
```bash
curl -X GET http://localhost:8080/api/v1/employees/999
```
**Response (404 Not Found):**
```json
{
  "error": {
    "code": "EMPLOYEE_NOT_FOUND",
    "message": "Employee not found"
  }
}
```

---

## 13. Known V1 Limitations

* **No Authentication / Authorization:** Endpoints are currently open without JWT/OAuth authentication (scheduled for security phase).
* **No Pagination & Filtering:** `GET /api/v1/employees` returns all records in a single payload.
* **Manual / Tool-Based Migrations:** Schema creation is version-controlled via SQL files in `migrations/` rather than automatically applied at application startup to avoid destructive changes in production.

---

## 14. Future Project Phases

The architecture is deliberately modular and decoupled to support future infrastructure and operational enhancements without application rewrites:

1. **Docker Containerization:** Multi-stage Docker build for minimal runtime image ([06-docker-design.md](06-docker-design.md))
2. **Local Orchestration:** Docker Compose environment pairing API and MySQL ([06-docker-design.md](06-docker-design.md))
3. **CI/CD Pipeline:** Jenkins automated testing, vetting, building, and publishing container images ([07-ci-cd-design.md](07-ci-cd-design.md))
4. **Kubernetes Deployment:** Deployments, Services, ConfigMaps, Secrets, and Rolling Updates ([08-kubernetes-design.md](08-kubernetes-design.md))
5. **AWS Cloud Production:** ECR, VPC, EKS Cluster, and Amazon RDS for MySQL ([09-aws-architecture.md](09-aws-architecture.md))
6. **Security & Observability:** JWT Auth, RBAC, TLS, AWS Secrets Manager, Prometheus, Grafana ([11-future-security-observability.md](11-future-security-observability.md))

---

## Reference Architecture Documentation

* [01 — Requirements](01-requirements.md)
* [02 — System Architecture](02-system-architecture.md)
* [03 — API Design](03-api-design.md)
* [04 — Database Design](04-database-design.md)
* [05 — Project Structure](05-project-structure.md)
* [06 — Docker Design](06-docker-design.md)
* [07 — CI/CD Design](07-ci-cd-design.md)
* [08 — Kubernetes Design](08-kubernetes-design.md)
* [09 — AWS Architecture](09-aws-architecture.md)
* [10 — Deployment Strategy](10-deployment-strategy.md)
* [11 — Future Security & Observability](11-future-security-observability.md)
