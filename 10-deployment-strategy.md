# Employee Management Platform — Deployment Strategy

## 1. Deployment Philosophy

Deployment will progress from the simplest environment to the target AWS environment.

The same application artifact should move through environments as consistently as possible.

## 2. Environment Progression

```text
Development
    ↓
Docker Compose
    ↓
Local Kubernetes
    ↓
AWS EKS
```

## 3. Stage 1 — Local Development

Run:

```text
Go API
+
MySQL
```

Validate all CRUD operations and tests.

## 4. Stage 2 — Docker Compose

Run:

```text
employee-api container
+
mysql container
```

Validate:

- Container startup.
- Service networking.
- Environment variables.
- Database connectivity.
- CRUD operations.

## 5. Stage 3 — GitHub

Commit source and infrastructure files.

Recommended repository flow:

```text
feature branch
     ↓
commit
     ↓
push
     ↓
pull request
     ↓
main
```

## 6. Stage 4 — Jenkins CI

Jenkins validates:

```text
Checkout
Test
Vet/static checks
Build
Docker build
Docker push
```

## 7. Stage 5 — Local Kubernetes

Deploy:

```text
Namespace
ConfigMap
Secret
Deployment
Service
```

Validate:

- Pods become Ready.
- Service routes traffic.
- Pods can connect to MySQL.
- Rolling update works.
- Rollback works.

## 8. Stage 6 — AWS Infrastructure

Create the AWS infrastructure in controlled steps:

```text
VPC
  ↓
Networking
  ↓
ECR
  ↓
EKS
  ↓
RDS
  ↓
Load Balancer
```

## 9. Stage 7 — AWS Deployment

Push image to ECR.

Deploy the image to EKS.

Configure the application to connect to RDS.

Verify:

```text
Internet
  ↓
Load Balancer
  ↓
EKS
  ↓
Employee API
  ↓
RDS MySQL
```

## 10. Stage 8 — CI/CD Deployment

Once manual deployment works:

```text
GitHub
  ↓
Jenkins
  ↓
Test
  ↓
Docker Build
  ↓
ECR
  ↓
EKS rollout
  ↓
Deployment verification
```

Automation should be added only after each manual operation is understood.

## 11. Rollback Strategy

At minimum, Kubernetes deployment rollback should be demonstrated.

Example concept:

```text
Version 1
   ↓
Version 2
   ↓
Problem detected
   ↓
Rollback
   ↓
Version 1
```

The exact rollback commands and pipeline behavior will be documented during implementation.

## 12. Verification

Every deployment stage must have verification.

Examples:

### Application

```text
GET /health
CRUD API tests
```

### Docker

```text
Container running
API reachable
Database reachable
```

### Kubernetes

```text
Pods Ready
Service reachable
Rollout successful
```

### AWS

```text
EKS workload Ready
Load Balancer reachable
RDS connectivity successful
CRUD operations successful
```

## 13. Cleanup

AWS resources should be explicitly tracked.

After learning sessions, unused resources should be removed where possible to control cost.

A cleanup checklist will be maintained during the AWS phase.
