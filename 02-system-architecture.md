# Employee Management Platform — System Architecture

## 1. Architecture Goal

The system uses a layered Go application and a progressively deployed infrastructure architecture.

The business application remains a modular monolith. Microservices are intentionally out of scope for the first version.

## 2. Application Architecture

```text
Client
  |
  v
Gin HTTP Router
  |
  v
Handler Layer
  |
  v
Service Layer
  |
  v
Repository Layer
  |
  v
MySQL
```

### Handler Layer

Responsibilities:

- Receive HTTP requests.
- Bind request data.
- Perform request-level validation.
- Call services.
- Translate application outcomes into HTTP responses.

The handler must not contain database queries or substantial business rules.

### Service Layer

Responsibilities:

- Business rules.
- Application orchestration.
- Transaction coordination where required.
- Translate repository outcomes into domain/application outcomes.

### Repository Layer

Responsibilities:

- Persistence.
- SQL/ORM operations.
- Query construction.
- Database-specific behavior.

### Model Layer

Contains domain/entity structures and appropriate request/response structures.

## 3. Infrastructure Architecture

### Local development

```text
Developer Machine
  |
  +-- Go application
  |
  +-- MySQL
  |
  +-- Docker Compose
        |
        +-- employee-api container
        +-- mysql container
```

### CI/CD

```text
Developer
   |
   v
GitHub
   |
   v
Jenkins
   |
   +-- Checkout
   +-- Test
   +-- Vet/static checks
   +-- Build
   +-- Docker Build
   +-- Docker Push
   |
   v
Container Registry
```

### Kubernetes

```text
Kubernetes Cluster
  |
  +-- Namespace
  |
  +-- Service
  |     |
  |     v
  +-- Deployment
        |
        +-- Pod
        +-- Pod
```

MySQL may run locally in Kubernetes for learning, but the target AWS architecture uses RDS MySQL rather than running the production database inside EKS.

## 4. AWS Target Architecture

```text
Internet
   |
   v
AWS Load Balancer
   |
   v
EKS
   |
   +-- Employee API Pods
   |
   v
RDS MySQL
```

AWS network components:

```text
AWS Region
  |
  v
VPC
  |
  +-- Public subnets
  |
  +-- Private subnets
       |
       +-- EKS worker/node infrastructure
       |
       +-- RDS MySQL
```

The exact subnet and routing design will be finalized during AWS implementation.

## 5. Image Flow

```text
Source Code
   |
   v
GitHub
   |
   v
Jenkins
   |
   v
Docker Build
   |
   v
Amazon ECR
   |
   v
EKS
```

Docker Hub can be used during the earlier learning phase if useful; ECR becomes the AWS-native registry for the EKS deployment.

## 6. Configuration Flow

Local:

```text
.env / environment
       |
       v
Docker Compose
       |
       v
Application
```

Kubernetes:

```text
ConfigMap / Secret
       |
       v
Pod environment
       |
       v
Application
```

AWS:

```text
Kubernetes configuration
       |
       v
EKS application
       |
       v
RDS
```

Advanced secrets management is deferred until the security phase.

## 7. Architectural Principles

- Stateless API containers.
- Database state is external to the API container.
- Configuration is externalized.
- Clear separation of concerns.
- Health endpoint available independently of business endpoints.
- Infrastructure is progressively introduced.
- Avoid premature microservices.
- Keep production concerns modular so security and observability can be added later.
