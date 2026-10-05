# Employee Management Platform — Requirements

## 1. Purpose

The Employee Management Platform is a production-oriented learning project whose business capability is intentionally simple: manage employees through a REST API.

The primary objective is to learn how a Go application is designed, containerized, tested, version-controlled, continuously integrated, deployed to Kubernetes, and ultimately deployed to AWS EKS with an AWS-managed MySQL database.

The project is separate from the existing `go-task-api` project.

## 2. Project Principles

- Create a completely new project and GitHub repository.
- Use Gemini CLI to generate and evolve the implementation.
- Documentation is the source of truth before implementation.
- Keep business functionality simple so infrastructure and operational learning remain the focus.
- Introduce infrastructure progressively rather than jumping directly to AWS.
- Keep security and observability out of the initial implementation; add them after the core platform is working.
- Use MySQL as the database locally and Amazon RDS for MySQL in AWS.
- Troubleshooting is handled in a separate error-handling chat; this chat remains the continuous project flow.

## 3. Functional Requirements

### Employee entity

An employee contains:

- `id`
- `first_name`
- `last_name`
- `email`
- `phone`
- `department`
- `designation`
- `salary`
- `joining_date`
- `created_at`
- `updated_at`

### CRUD operations

| Method | Endpoint | Purpose |
|---|---|---|
| POST | `/api/v1/employees` | Create an employee |
| GET | `/api/v1/employees` | List employees |
| GET | `/api/v1/employees/{id}` | Get one employee |
| PUT | `/api/v1/employees/{id}` | Update an employee |
| DELETE | `/api/v1/employees/{id}` | Delete an employee |

Future extensions may include pagination, filtering, sorting, authentication, authorization, and audit capabilities.

## 4. Validation Requirements

Initial validation should include:

- First name is required.
- Last name is required.
- Email is required and must have valid email syntax.
- Email must be unique.
- Phone is required and must conform to the application's defined format.
- Department is required.
- Designation is required.
- Salary must be non-negative.
- Joining date is required and must use a consistent date representation.

Validation failures must return a client-readable `400 Bad Request` response.

## 5. HTTP Behavior

Expected status codes:

- `200 OK` — successful read/update.
- `201 Created` — successful creation.
- `204 No Content` — successful deletion where no response body is needed.
- `400 Bad Request` — invalid request.
- `404 Not Found` — employee does not exist.
- `409 Conflict` — unique/conflict condition such as duplicate email.
- `500 Internal Server Error` — unexpected server-side failure.

The API must not expose raw database errors to clients.

## 6. Non-Functional Requirements

### Maintainability
Use clear package boundaries and separation of handler, service, repository, model, configuration, and database concerns.

### Testability
Business logic and repository behavior should be testable independently where practical.

### Configuration
Environment-specific configuration must not be hard-coded. Database credentials, connection information, and runtime configuration must be supplied through environment variables or platform configuration.

### Container readiness
The application must be able to run as a container and must not depend on files or tools existing only on the developer machine.

### Deployment readiness
The application must expose an HTTP health endpoint suitable for container and Kubernetes health checks.

### Graceful shutdown
The application should handle termination signals and shut down the HTTP server cleanly.

## 7. Initial Scope

In scope:

- Go
- Gin
- MySQL
- REST API
- CRUD
- Validation
- Error handling
- Automated tests
- Docker
- Docker Compose
- Git/GitHub
- Jenkins CI/CD
- Container registry
- Kubernetes
- AWS VPC
- Amazon ECR
- Amazon EKS
- Amazon RDS for MySQL

Deferred:

- Authentication
- Authorization/RBAC
- TLS/security hardening
- Secrets Manager integration
- Prometheus
- Grafana
- Centralized logging
- Distributed tracing
- Service mesh

## 8. Project Success Criteria

The project is considered successful when:

1. The application runs locally.
2. CRUD operations work against MySQL.
3. Tests pass.
4. The application runs using Docker Compose.
5. Source is maintained in GitHub.
6. Jenkins automatically tests and builds the application.
7. A versioned container image is published.
8. The application runs on Kubernetes.
9. The same application is deployed to EKS.
10. EKS can reach RDS MySQL.
11. The complete Git-to-EKS deployment flow works.
12. Security and observability can subsequently be added without redesigning the core application.
