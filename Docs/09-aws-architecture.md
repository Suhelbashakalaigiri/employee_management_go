# Employee Management Platform — AWS Architecture

## 1. AWS Objective

Move the working containerized/Kubernetes application to AWS and understand the underlying infrastructure rather than treating EKS as a black box.

## 2. Target AWS Components

```text
AWS Region
 |
 +-- IAM
 |
 +-- VPC
 |    |
 |    +-- Public Subnets
 |    +-- Private Subnets
 |    +-- Route Tables
 |    +-- Internet Gateway
 |    +-- NAT Gateway where required
 |    +-- Security Groups
 |
 +-- ECR
 |
 +-- EKS
 |
 +-- RDS MySQL
 |
 +-- Load Balancing
```

## 3. VPC

The VPC provides network isolation.

Conceptual structure:

```text
VPC
 |
 +-- Public Subnet(s)
 |
 +-- Private Subnet(s)
       |
       +-- EKS worker/node infrastructure
       |
       +-- RDS
```

The final subnet count and CIDR design will be selected during implementation based on AWS/EKS requirements and cost considerations.

## 4. IAM

IAM provides identities and permissions for AWS resources.

We will learn:

- Users/roles at a conceptual level.
- Policies.
- Least-privilege principles.
- EKS-related IAM requirements.
- How workloads and infrastructure authenticate to AWS.

Detailed security hardening is deferred.

## 5. ECR

Amazon ECR stores the application container image.

Flow:

```text
Jenkins
  |
  v
Docker Build
  |
  v
ECR
  |
  v
EKS
```

## 6. EKS

Amazon EKS hosts the Kubernetes application.

We will understand:

- Cluster.
- Control plane concept.
- Node groups.
- Networking.
- Kubernetes workloads.
- AWS integrations.

## 7. RDS

Amazon RDS for MySQL hosts persistent employee data.

The application in EKS connects to RDS over the AWS network.

```text
EKS Pod
  |
  v
AWS networking
  |
  v
RDS MySQL
```

## 8. Load Balancing

External API traffic will eventually be exposed through an AWS-managed load balancer.

The exact controller/service configuration will be selected during implementation.

## 9. Cost Awareness

Because this is a learning project using an AWS account with limited/free-tier considerations, resources must be created deliberately.

Before creating paid resources, we will verify:

- Current AWS pricing.
- Free-tier eligibility where applicable.
- Expected running cost.
- Cleanup procedure.

Resources that are not needed should be deleted after learning sessions.

## 10. AWS Implementation Order

```text
IAM fundamentals
   ↓
VPC
   ↓
Networking
   ↓
ECR
   ↓
EKS
   ↓
RDS MySQL
   ↓
Load Balancer
   ↓
Application deployment
   ↓
Jenkins → ECR → EKS
```

## 11. Deferred AWS Concerns

Later security/operations work may include:

- AWS Secrets Manager.
- KMS.
- Detailed IAM hardening.
- Private endpoints.
- Advanced network controls.
- CloudWatch integration.
- Monitoring and alerting.
