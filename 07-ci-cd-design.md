# Employee Management Platform — CI/CD Design

## 1. Objective

Jenkins will automate validation and container image creation so that the repository has a repeatable build process.

## 2. Initial Pipeline

```text
GitHub
  |
  v
Jenkins
  |
  +-- Checkout
  |
  +-- Test
  |
  +-- Vet / Static Checks
  |
  +-- Build
  |
  +-- Docker Build
  |
  +-- Docker Push
```

## 3. Pipeline Stages

### Checkout

Fetch source from GitHub.

### Test

Run the Go test suite.

Example:

```bash
go test ./...
```

### Vet / Static Checks

Run appropriate Go static checks.

### Build

Compile the application.

### Docker Build

Build the container image.

### Docker Push

Publish the image to the configured registry.

Initially Docker Hub may be used for learning. AWS deployment will use ECR.

## 4. Credentials

Credentials must be stored in Jenkins credentials management rather than in the Jenkinsfile.

Examples:

```text
Docker registry credentials
AWS credentials / role integration
GitHub credentials where required
```

Exact credential strategy will evolve with the AWS deployment.

## 5. Image Versioning

The pipeline should eventually produce immutable tags such as:

```text
employee-api:<git-sha>
```

A release tag may also be used:

```text
employee-api:1.0.0
```

## 6. Deployment Pipeline

After basic CI works, CD will be added:

```text
GitHub
  |
  v
Jenkins
  |
  +-- Test
  +-- Build
  +-- Docker Build
  +-- Push to ECR
  |
  v
Deploy to EKS
  |
  v
Verify rollout
```

Deployment automation will be introduced only after manual Kubernetes deployment is understood.

## 7. Failure Behavior

A failed stage must stop downstream stages.

For example:

```text
Test FAILED
   |
   X
Docker Build not executed
```

The pipeline should provide enough output to diagnose failures.

## 8. Jenkinsfile Principle

The Jenkinsfile should describe the repeatable pipeline, while secrets and environment-specific credentials remain outside source control.
