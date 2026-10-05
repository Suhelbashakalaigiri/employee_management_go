# Local Kubernetes Deployment — Employee Management Go

This document contains the complete command sequence used to deploy the Employee Management Go application to **Docker Desktop Kubernetes** locally.

The deployment includes:

- Go application
- MySQL
- PersistentVolumeClaim
- ConfigMap
- Secret
- Database migration Job
- Kubernetes Services
- Readiness/Liveness probes
- Resource requests/limits
- Local access using `kubectl port-forward`

---

# 1. Verify Kubernetes Configuration

## 1.1 Check available Kubernetes contexts

```powershell
kubectl config get-contexts
```

**Use case:**  
Shows all Kubernetes contexts configured on the machine and identifies which context is currently active.

---

## 1.2 Check the current Kubernetes context

```powershell
kubectl config current-context
```

**Use case:**  
Confirms which Kubernetes cluster `kubectl` commands will operate against.

Always check this before deploying, especially when you work with both local Kubernetes and EKS.

---

## 1.3 Switch to the required context

```powershell
kubectl config use-context <context-name>
```

Example:

```powershell
kubectl config use-context docker-desktop
```

**Use case:**  
Switches `kubectl` to a specific Kubernetes cluster.

> Use the context name returned by `kubectl config get-contexts`. Do not assume the name.

---

# 2. Verify the Kubernetes Cluster

## 2.1 Check cluster information

```powershell
kubectl cluster-info
```

**Use case:**  
Confirms that `kubectl` can communicate with the Kubernetes API server.

---

## 2.2 Check Kubernetes nodes

```powershell
kubectl get nodes
```

**Use case:**  
Checks whether Kubernetes nodes are available and ready.

Expected:

```text
NAME                    STATUS   ROLES           AGE   VERSION
desktop-control-plane   Ready    control-plane   ...   v1.36.1
```

The important value is:

```text
STATUS = Ready
```

---

## 2.3 Get detailed node information

```powershell
kubectl get nodes -o wide
```

**Use case:**  
Displays additional information such as:

- Internal IP
- External IP
- OS
- Kernel
- Container runtime

---

# 3. Validate Kubernetes Manifests Before Deployment

From the project root:

```powershell
kubectl apply --dry-run=client -f k8s/
```

**Use case:**  
Validates the Kubernetes manifests without actually creating resources.

Important:

```text
--dry-run=client
```

does **not** deploy anything.

It checks whether Kubernetes can understand the manifests.

---

# 4. Create ConfigMap

```powershell
kubectl apply -f k8s/configmap.yaml
```

**Use case:**  
Creates non-sensitive application configuration such as:

```text
DB_NAME
DB_PORT
APP_PORT
```

Verify:

```powershell
kubectl get configmap
```

---

# 5. Create Secret

```powershell
kubectl apply -f k8s/secret.yaml
```

**Use case:**  
Creates sensitive configuration such as the MySQL root password.

Verify:

```powershell
kubectl get secrets
```

Do not print Secret values unnecessarily.

---

# 6. Create MySQL PersistentVolumeClaim

```powershell
kubectl apply -f k8s/mysql-pvc.yaml
```

**Use case:**  
Requests persistent storage for MySQL.

Check:

```powershell
kubectl get pvc
```

Initially, the PVC may show:

```text
Pending
```

This is not necessarily an error.

Docker Desktop's StorageClass uses:

```text
WaitForFirstConsumer
```

so the PVC can remain Pending until a Pod actually consumes it.

---

# 7. Check StorageClass

```powershell
kubectl get storageclass
```

**Use case:**  
Shows the storage classes available in the cluster.

For Docker Desktop Kubernetes we used the default:

```text
standard
```

---

# 8. Inspect PVC When It Is Pending

```powershell
kubectl describe pvc mysql-pvc
```

**Use case:**  
Shows why a PVC is Pending or why storage provisioning failed.

Look at the `Events` section.

For:

```text
WaitForFirstConsumer
```

the expected behavior is:

```text
PVC created
    ↓
Waiting for Pod
    ↓
MySQL Pod scheduled
    ↓
Storage provisioned
    ↓
PVC becomes Bound
```

---

# 9. Deploy MySQL

```powershell
kubectl apply -f k8s/mysql-deployment.yaml
```

**Use case:**  
Creates the MySQL Deployment and MySQL Pod.

Check Pods:

```powershell
kubectl get pods
```

Check Deployment:

```powershell
kubectl get deployments
```

Check PVC:

```powershell
kubectl get pvc
```

Expected:

```text
mysql-pvc    Bound
```

MySQL Pod should eventually become:

```text
1/1    Running
```

---

# 10. Inspect MySQL Pod

Find the MySQL Pod:

```powershell
kubectl get pods
```

Then inspect it:

```powershell
kubectl describe pod <mysql-pod-name>
```

**Use case:**  
Useful for troubleshooting:

- Image problems
- Environment variables
- Volume mounting
- Probes
- Scheduling
- Container startup failures

---

# 11. Check MySQL Logs

```powershell
kubectl logs <mysql-pod-name>
```

**Use case:**  
Shows MySQL startup and runtime logs.

---

# 12. Create MySQL Service

```powershell
kubectl apply -f k8s/mysql-service.yaml
```

**Use case:**  
Creates a stable internal network endpoint for MySQL.

Check:

```powershell
kubectl get svc
```

Expected:

```text
employee-management-mysql   ClusterIP   ...   3306/TCP
```

The Go application connects to:

```text
employee-management-mysql:3306
```

rather than directly connecting to the MySQL Pod IP.

---

# 13. Check MySQL Service Endpoints

```powershell
kubectl get endpoints employee-management-mysql
```

**Use case:**  
Confirms that the MySQL Service has selected a MySQL Pod.

Example:

```text
employee-management-mysql   10.244.0.6:3306
```

Note: Kubernetes versions 1.33+ may display a deprecation warning for the old Endpoints API.

The modern alternative is:

```powershell
kubectl get endpointslice
```

---

# 14. Create Migration ConfigMap

```powershell
kubectl apply -f k8s/migration-configmap.yaml
```

**Use case:**  
Makes the SQL migration file available to the migration container.

Architecture:

```text
migration-configmap.yaml
        ↓
ConfigMap
        ↓
SQL migration file
        ↓
Migration container
```

Verify:

```powershell
kubectl get configmap
```

---

# 15. Run Database Migration Job

```powershell
kubectl apply -f k8s/migration.yaml
```

**Use case:**  
Creates a Kubernetes Job that executes the database migration.

The migration Job:

```text
Migration ConfigMap
        ↓
Migration container
        ↓
SQL migration
        ↓
MySQL
        ↓
employees table
```

---

# 16. Check Migration Job

```powershell
kubectl get jobs
```

Expected after successful migration:

```text
NAME                            STATUS     COMPLETIONS
employee-management-migration   Complete   1/1
```

---

# 17. Check Migration Pod

```powershell
kubectl get pods
```

The migration Pod may show:

```text
0/1   Completed
```

This is **normal**.

Unlike a Deployment, a Job is supposed to finish.

---

# 18. View Migration Logs

Find the migration Pod:

```powershell
kubectl get pods
```

Then:

```powershell
kubectl logs <migration-pod-name>
```

Example successful output:

```text
1/u create_employees (...)
```

This confirms that the migration executed successfully.

---

# 19. Important Migration Failure Scenario

If the migration Job fails before MySQL is ready, inspect:

```powershell
kubectl describe job employee-management-migration
```

If the Job reaches its `backoffLimit`, Kubernetes may delete the failed Pod.

To recreate the Job:

```powershell
kubectl delete job employee-management-migration
```

Then:

```powershell
kubectl apply -f k8s/migration.yaml
```

Check:

```powershell
kubectl get jobs
kubectl get pods
```

Then view logs:

```powershell
kubectl logs <migration-pod-name>
```

Important lesson:

```text
MySQL Service exists
        ≠
MySQL is necessarily ready
```

The migration container must be able to connect to an operational MySQL instance.

---

# 20. Deploy Go Application

After MySQL and migration are ready:

```powershell
kubectl apply -f k8s/go-app-deployment.yaml
```

**Use case:**  
Creates the Go application Deployment.

Our Deployment uses:

```text
replicas: 2
```

Therefore Kubernetes creates two Go Pods.

Check:

```powershell
kubectl get pods
```

Expected:

```text
employee-management-go-xxxxx   1/1   Running
employee-management-go-yyyyy   1/1   Running
```

---

# 21. Check Go Deployment

```powershell
kubectl get deployment
```

Or:

```powershell
kubectl get deployments
```

Expected:

```text
NAME                    READY
employee-management-go  2/2
```

---

# 22. Inspect Go Pod

```powershell
kubectl describe pod <go-pod-name>
```

**Use case:**  
Useful for checking:

- Container configuration
- Environment variables
- Resource requests/limits
- Probes
- Image
- Pod events

---

# 23. Check Go Application Logs

```powershell
kubectl logs <go-pod-name>
```

**Use case:**  
Checks Go application startup and runtime behavior.

---

# 24. Verify Health Inside Go Pod

```powershell
kubectl exec -it <go-pod-name> -- wget -qO- http://localhost:8080/health
```

Expected:

```json
{"status":"UP"}
```

This proves:

```text
Go Pod
   ↓
Go application
   ↓
8080
   ↓
/health
```

is working.

---

# 25. Create Go Application Service

```powershell
kubectl apply -f k8s/go-app-service.yaml
```

**Use case:**  
Creates a stable Kubernetes endpoint for the Go application.

Our local Service is:

```text
Type: NodePort
Port: 8080
NodePort: 30080
```

Check:

```powershell
kubectl get svc
```

Expected:

```text
employee-management-go   NodePort   ...   8080:30080/TCP
```

---

# 26. Check Go Service Endpoints

```powershell
kubectl get endpoints employee-management-go
```

**Use case:**  
Confirms that the Service has selected the Go Pods.

The Service selector:

```yaml
selector:
  app: employee-management-go
```

matches the Pod labels:

```yaml
labels:
  app: employee-management-go
```

Therefore:

```text
Go Service
    ↓
selector
    ↓
Go Pod 1
Go Pod 2
```

---

# 27. Verify Kubernetes Service Internally

Create a temporary curl Pod:

```powershell
kubectl run test-curl --rm -it --image=curlimages/curl -- sh
```

Inside the temporary Pod:

```sh
curl http://employee-management-go:8080/health
```

Expected:

```json
{"status":"UP"}
```

Exit:

```sh
exit
```

The temporary Pod is automatically removed because of:

```text
--rm
```

This verifies:

```text
Temporary Pod
      ↓
Go Service
      ↓
Go Pod
      ↓
Go application
```

---

# 28. Local Port Forwarding

Docker Desktop Kubernetes NodePort access from Windows may not work directly depending on the local Docker Desktop/WSL2 networking configuration.

In our environment, the application was healthy internally but:

```text
localhost:30080
```

was not reachable from Windows/Postman.

Use:

```powershell
kubectl port-forward service/employee-management-go 8081:8080
```

Expected:

```text
Forwarding from 127.0.0.1:8081 -> 8080
Forwarding from [::1]:8081 -> 8080
```

Keep this terminal running.

---

# 29. Test Health From Postman

Use:

```text
GET http://localhost:8081/health
```

Expected:

```json
{
  "status": "UP"
}
```

---

# 30. Test Create Employee From Postman

Use:

```text
POST http://localhost:8081/api/v1/employees
```

Header:

```text
Content-Type: application/json
```

Body:

```json
{
  "first_name": "Rahul",
  "last_name": "Kumar",
  "email": "rahul.kumar@example.com",
  "phone": "+919876543210",
  "department": "Engineering",
  "designation": "Software Engineer",
  "salary": 65000,
  "joining_date": "2026-09-24"
}
```

---

# 31. Test Other CRUD APIs

## Get all employees

```text
GET http://localhost:8081/api/v1/employees
```

## Get employee

```text
GET http://localhost:8081/api/v1/employees/1
```

## Update employee

```text
PUT http://localhost:8081/api/v1/employees/1
```

## Delete employee

```text
DELETE http://localhost:8081/api/v1/employees/1
```

The Kubernetes path is now:

```text
Postman
   ↓
localhost:8081
   ↓
kubectl port-forward
   ↓
Go Service :8080
   ↓
Go Pod
   ↓
Go application
   ↓
MySQL Service :3306
   ↓
MySQL Pod
   ↓
PersistentVolume
```

---

# 32. Useful Monitoring Commands

## View all Pods

```powershell
kubectl get pods
```

## Watch Pods continuously

```powershell
kubectl get pods -w
```

**Use case:**  
Useful when waiting for Pods to start, restart, or become Ready.

---

## View all Services

```powershell
kubectl get svc
```

---

## View all Deployments

```powershell
kubectl get deployments
```

---

## View all Jobs

```powershell
kubectl get jobs
```

---

## View all PVCs

```powershell
kubectl get pvc
```

---

## View all ConfigMaps

```powershell
kubectl get configmaps
```

---

## View all Secrets

```powershell
kubectl get secrets
```

---

## View everything in the default namespace

```powershell
kubectl get all
```

Note that `get all` does not literally mean every Kubernetes resource type; it shows the common workload/service resources.

---

# 33. Troubleshooting Commands

## Describe a Pod

```powershell
kubectl describe pod <pod-name>
```

Use when a Pod is:

```text
Pending
ContainerCreating
CrashLoopBackOff
Error
NotReady
```

---

## View Pod logs

```powershell
kubectl logs <pod-name>
```

---

## Follow Pod logs

```powershell
kubectl logs -f <pod-name>
```

---

## View logs from a previous crashed container

```powershell
kubectl logs <pod-name> --previous
```

---

## Describe Deployment

```powershell
kubectl describe deployment <deployment-name>
```

---

## Describe Service

```powershell
kubectl describe service <service-name>
```

---

## Describe Job

```powershell
kubectl describe job <job-name>
```

---

## Describe PVC

```powershell
kubectl describe pvc <pvc-name>
```

---

# 34. Check Current Cluster Resources

A useful complete check:

```powershell
kubectl get pods
kubectl get deployments
kubectl get svc
kubectl get jobs
kubectl get pvc
kubectl get configmaps
kubectl get secrets
```

Expected overall state:

```text
Go Pods                    Running
MySQL Pod                  Running
Migration Job              Complete
PVC                        Bound
Go Service                 NodePort
MySQL Service              ClusterIP
ConfigMap                  Created
Secret                     Created
```

---

# 35. Final Local Kubernetes Architecture

```text
                         Windows / Postman
                                │
                                │ localhost:8081
                                ▼
                       kubectl port-forward
                                │
                                ▼
                    Go Kubernetes Service
                         :8080 / ClusterIP
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
                 Go Pod 1                Go Pod 2
                 :8080                   :8080
                    │                       │
                    └───────────┬───────────┘
                                │
                                ▼
                     MySQL Kubernetes Service
                              :3306
                                │
                                ▼
                           MySQL Pod
                              │
                              ▼
                          mysql-pvc
                              │
                              ▼
                       Persistent Storage


Migration flow:

migration-configmap
        │
        ▼
migration Job
        │
        ▼
MySQL Service
        │
        ▼
employee_management
        │
        ▼
employees table
```

---

# 36. Important Kubernetes Concepts Demonstrated

This deployment demonstrated:

- Kubernetes Context
- Kubernetes Cluster
- Node
- Deployment
- ReplicaSet
- Pod
- Service
- ClusterIP
- NodePort
- ConfigMap
- Secret
- PersistentVolumeClaim
- StorageClass
- Dynamic volume provisioning
- `WaitForFirstConsumer`
- Kubernetes Job
- Container environment variables
- Service discovery
- Kubernetes DNS
- Labels and selectors
- Readiness probes
- Liveness probes
- Resource requests
- Resource limits
- Migration execution
- `kubectl exec`
- `kubectl logs`
- `kubectl describe`
- `kubectl port-forward`

---

# 37. Key Lessons

### Pod IP should not be used by the application

Use the Service:

```text
employee-management-mysql:3306
```

instead of:

```text
10.244.x.x:3306
```

Pod IPs are ephemeral.

### Service selects Pods using labels

```yaml
selector:
  app: employee-management-go
```

matches:

```yaml
labels:
  app: employee-management-go
```

### ConfigMap is for normal configuration

Example:

```text
DB_NAME
DB_PORT
APP_PORT
```

### Secret is for sensitive configuration

Example:

```text
root-password
```

### PVC provides persistent storage

MySQL data is stored under:

```text
/var/lib/mysql
```

and backed by:

```text
mysql-pvc
```

### Deployment keeps applications running

Go:

```text
replicas: 2
```

### Job performs a finite task

Migration:

```text
Run migration
    ↓
Complete
    ↓
Stop
```

### Port forwarding is for local development/debugging

```powershell
kubectl port-forward service/employee-management-go 8081:8080
```

It is not the production exposure mechanism.

---

# 38. Cleanup Commands

If you want to remove the complete local deployment:

```powershell
kubectl delete -f k8s/
```

Check:

```powershell
kubectl get all
kubectl get pvc
```

If you also want to remove the MySQL persistent data:

```powershell
kubectl delete pvc mysql-pvc
```

**Warning:** deleting the PVC can delete the locally stored MySQL data depending on the StorageClass reclaim policy.

---

# 39. Complete Deployment Command Sequence

For future reference, the practical deployment sequence is:

```powershell
# 1. Verify context
kubectl config get-contexts
kubectl config current-context

# 2. Verify cluster
kubectl cluster-info
kubectl get nodes

# 3. Validate manifests
kubectl apply --dry-run=client -f k8s/

# 4. Configuration
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# 5. Storage
kubectl apply -f k8s/mysql-pvc.yaml
kubectl get pvc

# 6. MySQL
kubectl apply -f k8s/mysql-deployment.yaml
kubectl get pods
kubectl get pvc

# 7. MySQL Service
kubectl apply -f k8s/mysql-service.yaml
kubectl get svc

# 8. Migration configuration
kubectl apply -f k8s/migration-configmap.yaml

# 9. Database migration
kubectl apply -f k8s/migration.yaml
kubectl get jobs
kubectl get pods

# 10. Verify migration logs
kubectl logs <migration-pod-name>

# 11. Go application
kubectl apply -f k8s/go-app-deployment.yaml
kubectl get pods
kubectl get deployments

# 12. Go Service
kubectl apply -f k8s/go-app-service.yaml
kubectl get svc

# 13. Verify Service → Pods
kubectl get endpoints employee-management-go

# 14. Verify application internally
kubectl exec -it <go-pod-name> -- wget -qO- http://localhost:8080/health

# 15. Test Service internally
kubectl run test-curl --rm -it --image=curlimages/curl -- sh

# Inside temporary Pod:
curl http://employee-management-go:8080/health
exit

# 16. Expose locally to Windows/Postman
kubectl port-forward service/employee-management-go 8081:8080

# 17. Postman
GET http://localhost:8081/health
POST http://localhost:8081/api/v1/employees
GET http://localhost:8081/api/v1/employees
GET http://localhost:8081/api/v1/employees/1
PUT http://localhost:8081/api/v1/employees/1
DELETE http://localhost:8081/api/v1/employees/1
```

---

# 40. Final Status

Local Kubernetes deployment is complete.

```text
Go Application                    ✅
Docker Image                      ✅
Docker Desktop Kubernetes         ✅
ConfigMap                         ✅
Secret                            ✅
MySQL Deployment                  ✅
MySQL Service                     ✅
PersistentVolumeClaim             ✅
Migration ConfigMap               ✅
Migration Job                     ✅
Go Deployment                     ✅
Go Service                        ✅
2 Go replicas                     ✅
MySQL persistent storage          ✅
Readiness/Liveness probes         ✅
Internal Kubernetes networking    ✅
CRUD through Kubernetes           ✅
Local Postman access              ✅
```

The next major stage is **AWS/EKS deployment**. Jenkins CI/CD can be introduced afterward, using the Kubernetes deployment model established here as the foundation.