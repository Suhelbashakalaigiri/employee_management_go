# Employee Management Platform — Docker Design

## 1. Goals

Docker provides a reproducible runtime for the Go API and local infrastructure.

The container image must:

- Build deterministically.
- Contain only runtime requirements.
- Avoid unnecessary tooling.
- Support environment-based configuration.
- Run the API as a non-development runtime process.

## 2. Multi-Stage Build

The preferred design is a multi-stage Docker build:

```text
Builder Image
    |
    +-- download dependencies
    +-- compile Go binary
    |
    v
Runtime Image
    |
    +-- compiled binary
    +-- required runtime files
    |
    v
Container
```

The exact base images will be selected during implementation.

## 3. Image Tagging

Avoid relying only on:

```text
latest
```

Preferred release-oriented tags include:

```text
employee-api:1.0.0
employee-api:<git-sha>
```

A Git SHA tag makes an image traceable to source code.

## 4. Docker Compose

Local Compose environment:

```text
docker-compose
   |
   +-- employee-api
   |
   +-- mysql
```

The two services share a private Compose network.

The API uses:

```text
mysql:<port>
```

where `mysql` is the Compose service name.

## 5. Environment Configuration

The image must not embed environment-specific credentials.

Example variables:

```text
APP_PORT
DB_HOST
DB_PORT
DB_NAME
DB_USER
DB_PASSWORD
```

`.env.example` documents required variables without containing real credentials.

## 6. Health

The application exposes:

```text
GET /health
```

This endpoint can later be used for Docker and Kubernetes health checks.

## 7. Container Lifecycle

The application should:

- Start predictably.
- Log startup failures clearly.
- Handle SIGTERM.
- Close database resources during shutdown.
- Exit with an appropriate non-zero status when startup cannot succeed.

## 8. Docker Security

Full container security hardening is deferred to the later security phase.

The initial implementation should still avoid obvious anti-patterns such as embedding credentials in the Dockerfile.
