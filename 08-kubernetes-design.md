# Employee Management Platform — Kubernetes Design

## 1. Goal

Deploy the stateless Go API as a Kubernetes workload and learn Kubernetes operations before moving to EKS.

## 2. Core Resources

Initial resources:

```text
Namespace
Deployment
Service
ConfigMap
Secret
```

Later:

```text
HorizontalPodAutoscaler
Ingress / LoadBalancer integration
PodDisruptionBudget where justified
```

## 3. Deployment

The API runs as a Deployment.

Initial conceptual design:

```text
Deployment
   |
   +-- Pod
   +-- Pod
```

Multiple replicas are used to demonstrate rolling deployment and service load balancing.

## 4. Service

A Kubernetes Service provides a stable network endpoint for the pods.

```text
Client
  |
  v
Service
  |
  +-- Pod
  +-- Pod
```

## 5. Configuration

Non-sensitive configuration:

```text
ConfigMap
```

Sensitive configuration:

```text
Secret
```

Initial secrets use Kubernetes Secret resources for learning. More advanced secret management is deferred to the security phase.

## 6. Health Checks

The application should expose `/health`.

Kubernetes should eventually use:

```text
livenessProbe
readinessProbe
```

The probes have different responsibilities:

- Liveness: whether the process should be restarted.
- Readiness: whether the pod should receive traffic.

## 7. Resource Management

The Deployment should define CPU and memory requests and limits after application behavior is measured sufficiently to choose reasonable initial values.

## 8. Rolling Updates

Kubernetes should perform rolling updates so that new versions can replace old pods without unnecessary downtime.

The deployment process should support rollback.

## 9. Local Kubernetes

Before EKS, the application should be deployed to a local Kubernetes environment.

This provides a safe environment to learn:

```bash
kubectl get
kubectl describe
kubectl logs
kubectl exec
kubectl rollout
kubectl apply
kubectl delete
```

## 10. AWS Kubernetes

The final target is Amazon EKS.

The EKS architecture will separate:

```text
Internet-facing traffic
        |
        v
AWS Load Balancer
        |
        v
EKS Service
        |
        v
Employee API Pods
        |
        v
RDS MySQL
```

## 11. Database Placement

For AWS:

```text
MySQL → RDS
```

The production database is not run as a normal MySQL pod inside EKS.

## 12. Helm

Raw manifests should be understood first.

Helm can then be introduced to package the Kubernetes deployment and manage environment-specific values.

