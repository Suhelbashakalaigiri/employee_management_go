# AWS EKS Deployment Guide — Employee Management Go

This document is the complete guide and command sequence for deploying the **Employee Management Go** application to **Amazon Elastic Kubernetes Service (EKS)** in the `ap-south-1` (Mumbai) region running Kubernetes **1.36**.

---

## 1. Directory Structure & Manifest Purpose

All EKS-specific manifests are stored in the isolated directory:
`k8s/eks/`

```text
k8s/eks/
├── configmap.yaml            # Non-sensitive application & DB configuration (DB_NAME, DB_PORT, APP_PORT)
├── secret-example.yaml       # Template for sensitive credentials (placeholder only, root-password)
├── storageclass.yaml         # AWS EBS gp3 dynamic volume provisioner (ebs.csi.aws.com)
├── mysql-pvc.yaml            # PersistentVolumeClaim (5Gi, ReadWriteOnce, ebs-sc StorageClass)
├── mysql-deployment.yaml     # MySQL 8.0 single-replica Deployment with persistent EBS mount & readiness probe
├── mysql-service.yaml        # Internal ClusterIP Service exposing MySQL on port 3306 (DB_HOST)
├── migration-configmap.yaml  # ConfigMap containing 000001_create_employees.up.sql DDL script
├── migration.yaml            # Kubernetes Job running migrate/migrate to initialize the database schema
├── go-app-deployment.yaml    # Go application Deployment (2 replicas, CPU/memory limits, readiness/liveness probes)
├── go-app-service.yaml       # AWS LoadBalancer Service exposing the Go API on port 8080 for external access
└── EKS-DEPLOYMENT.md         # This comprehensive operational and deployment guide
```

### Detailed Purpose of Each File

| Manifest File | Kind | Purpose |
| :--- | :--- | :--- |
| `configmap.yaml` | `ConfigMap` | Stores environment values `DB_NAME` (`employee_management`), `DB_PORT` (`3306`), and `APP_PORT` (`8080`). Injected into MySQL, migration Job, and Go app. |
| `secret-example.yaml` | `Secret` | Sample template with placeholder `root-password: CHANGE_ME`. **Never committed with real credentials.** The real secret is created manually via `kubectl`. |
| `storageclass.yaml` | `StorageClass` | Defines `ebs-sc` using `ebs.csi.aws.com` and `gp3`. Uses `WaitForFirstConsumer` so AWS EBS volumes are provisioned in the exact Availability Zone where MySQL is scheduled. |
| `mysql-pvc.yaml` | `PersistentVolumeClaim` | Requests 5Gi of persistent block storage bound to `ebs-sc`. Dynamically provisions an AWS EBS gp3 volume. |
| `mysql-deployment.yaml` | `Deployment` | Runs container `mysql:8.0` with 1 replica, mounts the PVC at `/var/lib/mysql`, and runs `mysqladmin ping` as readiness probe. |
| `mysql-service.yaml` | `Service` | Creates internal `ClusterIP` DNS name `employee-management-mysql:3306` accessible within the cluster. |
| `migration-configmap.yaml` | `ConfigMap` | Mounts `000001_create_employees.up.sql` as a volume file for the migration tool. |
| `migration.yaml` | `Job` | One-off Batch Job running `migrate/migrate:v4.19.0` to apply DDL migrations to MySQL before the Go app starts. |
| `go-app-deployment.yaml` | `Deployment` | Runs `suhelbasha7324/employee-management-go:1.0` with 2 replicas, connects to `employee-management-mysql`, configured with probes on `/health`. |
| `go-app-service.yaml` | `Service` | Type `LoadBalancer`. Triggers AWS to provision a public Load Balancer so external clients (Postman, browser) can reach port 8080. |

---

## 2. EKS Architecture

The target EKS architecture runs in **AWS Region `ap-south-1`** using Kubernetes version **1.36**:

```text
                                       +-------------------------------------------------------------+
                                       |                      AWS Region: ap-south-1                 |
                                       |                                                             |
   [ Postman / Browser ]               |  +-------------------------------------------------------+  |
             |                         |  |                   Amazon EKS Cluster                  |  |
             | HTTP :8080              |  |              (Kubernetes Control Plane v1.36)         |  |
             v                         |  +-------------------------------------------------------+  |
+--------------------------+           |                                                             |
|    AWS Load Balancer     | --------> |  +-------------------------------------------------------+  |
| (Classic / Network LB)   |           |  |                  EKS Managed Node Group               |  |
+--------------------------+           |  |                                                       |  |
             |                         |  |   +------------------------------------------------+  |  |
             |                         |  |   | Service: employee-management-go (LoadBalancer) |  |  |
             |                         |  |   | Port: 8080 -> TargetPort: 8080                 |  |  |
             |                         |  |   +-----------------------+------------------------+  |  |
             |                         |  |                           |                           |  |
             +-------------------------+----> Pod 1: Go App (8080)   +---> Pod 2: Go App (8080)   |  |
                                       |            |                             |               |  |
                                       |            +--------------+--------------+               |  |
                                       |                           |                              |  |
                                       |                           v                              |  |
                                       |        Service: employee-management-mysql (ClusterIP)    |  |
                                       |                           |                              |  |
                                       |                           v                              |  |
                                       |             Pod: employee-management-mysql (3306)        |  |
                                       |                           |                              |  |
                                       |                           v                              |  |
                                       |                   PVC: mysql-pvc                         |  |
                                       |                           |                              |  |
                                       |                           v                              |  |
                                       |               StorageClass: ebs-sc (gp3)                 |  |
                                       |  +------------------------+------------------------------+  |
                                       +---------------------------|---------------------------------+
                                                                   v
                                                  +----------------------------------+
                                                  |        AWS EBS gp3 Volume        |
                                                  |    (5Gi Persistent Block Store)  |
                                                  +----------------------------------+
```

### Architectural Key Points
1. **Control Plane:** AWS-managed EKS Control Plane (Kubernetes 1.36) deployed across multiple AZs in `ap-south-1`.
2. **Worker Nodes:** EC2 instances running in the EKS Managed Node Group with EBS CSI Driver daemonset / controller installed.
3. **External Traffic Flow:** Outside traffic enters through the AWS Load Balancer on port 8080, route traffic to healthy Go application pods.
4. **Internal Workload Communication:** Go pods query MySQL via the internal Kubernetes DNS name `employee-management-mysql:3306`.
5. **Decoupled Database Migration:** Migration runs as an isolated Kubernetes Job before the Go Deployment starts.

---

## 3. How EKS Manifests Differ From Local Kubernetes

| Dimension | Local Kubernetes (Docker Desktop) | AWS EKS | Rationale |
| :--- | :--- | :--- | :--- |
| **StorageClass** | `standard` (hostpath / local disk) | `ebs-sc` (`ebs.csi.aws.com` / `gp3`) | EKS nodes are EC2 instances. Local hostpath storage is lost when nodes terminate. EBS provides persistent block storage across pod restarts. |
| **Volume Binding** | `Immediate` or local default | `WaitForFirstConsumer` | In multi-AZ AWS environments, an EBS volume must be created in the exact Availability Zone where the MySQL pod is scheduled. |
| **Reclaim Policy** | Default (often Delete or hostpath) | `Delete` (configured in `ebs-sc`) | Prevents orphaned EBS volumes from continuously accumulating AWS billing charges when testing in dev/learning environments. |
| **Go Service Type** | `NodePort` (port 30080) | `LoadBalancer` (port 8080) | `NodePort` requires direct access to EC2 node IPs and firewall adjustments. `LoadBalancer` automatically provisions a public AWS Load Balancer for direct access from Postman. |
| **Secret Management** | Local plain `secret.yaml` with test password | Manual creation via `kubectl`, `secret-example.yaml` template | Prevents leaking actual database credentials into Git source control. |
| **Replica Scale** | 2 replicas tested on single local node | 2 replicas distributed across cloud worker nodes | Demonstrates cloud pod distribution, load balancing, and failure isolation. |

---

## 4. Why EBS Is Used for MySQL

In Kubernetes, pods are ephemeral by design. When a pod is deleted, rescheduled, or upgraded, its local filesystem is destroyed.

1. **Persistent Block Storage:** AWS Elastic Block Store (gp3) provides reliable, network-attached block storage that exists independently of the EC2 instance lifetime.
2. **Crash & Node Rescheduling Recovery:** If the EC2 worker node hosting MySQL fails, Kubernetes detects the node failure and reschedules the MySQL pod on another healthy worker node. The AWS EBS CSI driver unmounts the EBS volume from the failed node and mounts it to the new node, preserving all database data.
3. **ReadWriteOnce (RWO):** MySQL requires single-writer exclusive access to database files (`/var/lib/mysql`) to ensure ACID compliance and prevent data corruption. EBS volumes attach natively as `ReadWriteOnce`.
4. **Learning Milestone:** In production enterprise environments, managed databases like **Amazon RDS MySQL** are often preferred for automated backups and Multi-AZ replication. However, running MySQL with EBS in EKS is a critical learning exercise to understand Kubernetes CSI storage, PVC lifecycles, and volume binding.

---

## 5. Why LoadBalancer Is Used Instead of NodePort

1. **Accessibility from Windows / Postman:**
   - On Docker Desktop, NodePort often encounters WSL2 networking limitations, requiring `kubectl port-forward`.
   - On EKS, worker nodes typically sit in private subnets or behind security groups. NodePort (e.g., `:30080`) is not accessible over the public internet unless nodes have public IPs and security groups are modified.
2. **Cloud Integration:**
   - Specifying `type: LoadBalancer` signals the AWS Cloud Controller Manager to provision an external AWS Load Balancer (CLB or NLB) with a public DNS hostname.
   - Postman and API clients can send requests directly to `http://<load-balancer-dns>:8080` without any manual port forwarding or SSH tunnels.
3. **High Availability & Health Checks:**
   - The AWS Load Balancer automatically monitors worker node health and distributes incoming traffic between both Go application replicas.

---

## 6. Required Prerequisites Before Deployment

Before applying these manifests, ensure the following steps have been completed:

### Prerequisite 1: AWS EKS Cluster Ready
- EKS cluster created manually in AWS Console:
  - **Region:** `ap-south-1` (Mumbai)
  - **Kubernetes Version:** `1.36`
  - **Cluster Name:** (e.g., `employee-management-eks`)
  - **Worker Node Group:** At least 2 worker nodes (e.g., `t3.medium` instances) with status `Active`.

### Prerequisite 2: Amazon EBS CSI Driver Installed
AWS EKS does not include the EBS CSI driver by default. It must be enabled:
1. In the AWS Console, open the EKS Cluster -> **Add-ons** tab -> **Get more add-ons**.
2. Select **Amazon EBS CSI Driver**.
3. Ensure the worker node IAM role or the CSI Driver service account has the AWS-managed policy attached:
   `AmazonEBSCSIDriverPolicy`
4. Confirm the add-on status is `Active`.

### Prerequisite 3: Local Tools & `kubectl` Context Configured
Ensure `aws` CLI and `kubectl` are installed on your machine.
Update your local kubeconfig to point to your new EKS cluster:

```powershell
# Authenticate and update kubeconfig for EKS
aws eks update-kubeconfig --region ap-south-1 --name <your-eks-cluster-name>
```

Verify your active context:

```powershell
kubectl config current-context
```

It should return an ARN similar to:
`arn:aws:eks:ap-south-1:<account-id>:cluster/<your-eks-cluster-name>`

---

## 7. How to Create the Real Secret Manually

Never commit plaintext passwords to Git. The file `k8s/eks/secret-example.yaml` is only a reference.

Create the real Kubernetes Secret directly using `kubectl`:

```powershell
# Replace 'YourSecurePassword123!' with your desired MySQL root password
kubectl create secret generic mysql-secret --from-literal=root-password='YourSecurePassword123!'
```

Verify the secret was created (this does NOT print the password value):

```powershell
kubectl get secret mysql-secret
```

Expected output:
```text
NAME           TYPE     DATA   AGE
mysql-secret   Opaque   1      5s
```

---

## 8. Ordered Deployment Commands

Execute the manifests in this exact sequence to ensure all dependencies are satisfied:

```text
1. Verify Cluster & Nodes
           ↓
2. Verify EBS CSI Driver
           ↓
3. Create ConfigMap & Secret
           ↓
4. Create StorageClass & PVC
           ↓
5. Deploy MySQL & MySQL Service
           ↓
6. Wait for MySQL Pod Ready
           ↓
7. Run Migration Job
           ↓
8. Deploy Go Application & LoadBalancer Service
           ↓
9. Verify External LoadBalancer & Test Endpoints
```

### Step 1: Verify EKS Cluster Communication
```powershell
kubectl cluster-info
```

### Step 2: Verify Worker Nodes are Ready
```powershell
kubectl get nodes -o wide
```
*Wait until all nodes show `STATUS: Ready`.*

### Step 3: Verify EBS CSI Driver DaemonSet
```powershell
kubectl get pods -n kube-system -l app.kubernetes.io/name=aws-ebs-csi-driver
```
*Ensure the CSI controller and node daemonset pods are Running.*

### Step 4: Create ConfigMap
```powershell
kubectl apply -f k8s/eks/configmap.yaml
```

### Step 5: Create Secret Manually (if not already created in Section 7)
```powershell
kubectl create secret generic mysql-secret --from-literal=root-password='YourSecurePassword123!'
```

### Step 6: Create StorageClass
```powershell
kubectl apply -f k8s/eks/storageclass.yaml
```

### Step 7: Create PersistentVolumeClaim
```powershell
kubectl apply -f k8s/eks/mysql-pvc.yaml
```
*Note: Due to `WaitForFirstConsumer`, the PVC may stay in `Pending` status until the MySQL pod is scheduled. This is expected.*

### Step 8: Deploy MySQL
```powershell
kubectl apply -f k8s/eks/mysql-deployment.yaml
```

### Step 9: Create MySQL Service
```powershell
kubectl apply -f k8s/eks/mysql-service.yaml
```

### Step 10: Wait for MySQL to Become Ready
```powershell
kubectl wait --for=condition=ready pod -l app=employee-management-mysql --timeout=180s
```
Verify the PVC is now `Bound`:
```powershell
kubectl get pvc mysql-pvc
```

### Step 11: Create Migration ConfigMap
```powershell
kubectl apply -f k8s/eks/migration-configmap.yaml
```

### Step 12: Run Database Migration Job
```powershell
kubectl apply -f k8s/eks/migration.yaml
```

### Step 13: Verify Migration Completed
```powershell
kubectl wait --for=condition=complete job/employee-management-migration --timeout=120s
kubectl logs job/employee-management-migration
```
*Expected log output: `1/u create_employees (...)`.*

### Step 14: Deploy Go Application
```powershell
kubectl apply -f k8s/eks/go-app-deployment.yaml
```
Wait for both replicas to become Ready:
```powershell
kubectl wait --for=condition=ready pod -l app=employee-management-go --timeout=120s
```

### Step 15: Create Go LoadBalancer Service
```powershell
kubectl apply -f k8s/eks/go-app-service.yaml
```

### Step 16: Wait for External Load Balancer Provisioning
```powershell
kubectl get svc employee-management-go -w
```
*Press Ctrl+C once the `EXTERNAL-IP` column changes from `<pending>` to an AWS DNS address (e.g., `a1b2c3...ap-south-1.elb.amazonaws.com`).*

### Step 17: Test Health Endpoint
```powershell
# Retrieve the external hostname
$LB_HOST = (kubectl get svc employee-management-go -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl "http://${LB_HOST}:8080/health"
```

### Step 18: Test CRUD APIs
Proceed to Section 12 to run the full Postman CRUD testing sequence.

---

## 9. Comprehensive Verification Commands

Run these commands to inspect all cluster resources:

### Check Nodes
```powershell
kubectl get nodes -o wide
kubectl describe node <node-name>
```

### Check Pods
```powershell
kubectl get pods -o wide
```

### Check Deployments
```powershell
kubectl get deployments -o wide
```

### Check Services
```powershell
kubectl get svc -o wide
```

### Check PVC & PV
```powershell
kubectl get pvc
kubectl get pv
kubectl describe pvc mysql-pvc
```

### Check StorageClass
```powershell
kubectl get storageclass
kubectl describe storageclass ebs-sc
```

### Check Migration Job
```powershell
kubectl get jobs
kubectl describe job employee-management-migration
```

### Check Application & Database Logs
```powershell
# View MySQL logs
kubectl logs -l app=employee-management-mysql --tail=50

# View Go application logs (all pods)
kubectl logs -l app=employee-management-go --tail=50

# Follow logs from a specific Go pod
kubectl logs -f <go-pod-name>
```

---

## 10. How to Test the `/health` Endpoint

### Option A: Via AWS External Load Balancer (Primary Method)
```powershell
$LB_DNS = (kubectl get svc employee-management-go -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl "http://${LB_DNS}:8080/health"
```
Expected output:
```json
{"status":"UP"}
```

### Option B: Via `kubectl exec` Directly Inside Pod (Internal Verification)
```powershell
$POD_NAME = (kubectl get pods -l app=employee-management-go -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POD_NAME -- wget -qO- http://localhost:8080/health
```

### Option C: Via `kubectl port-forward` (Fallback/Debugging Method)
```powershell
kubectl port-forward svc/employee-management-go 8080:8080
```
Then in another terminal or browser:
```powershell
curl http://localhost:8080/health
```

---

## 11. How to Retrieve the External LoadBalancer Address

### In PowerShell (Windows)
```powershell
$LB_HOST = (kubectl get svc employee-management-go -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
Write-Host "External LoadBalancer URL: http://${LB_HOST}:8080"
```

### In Bash / Linux / macOS
```bash
LB_HOST=$(kubectl get svc employee-management-go -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
echo "External LoadBalancer URL: http://${LB_HOST}:8080"
```

> **Note on AWS DNS Propagation:** Newly provisioned AWS Load Balancers take approximately 2–3 minutes for AWS DNS records to propagate globally. If you receive a `Could not resolve host` error immediately after creation, wait 2 minutes and retry.

---

## 12. Postman Testing Sequence

Configure Postman with the variable:
- `baseUrl`: `http://<YOUR_LOAD_BALANCER_DNS>:8080`

### 1. Health Check
- **Method:** `GET`
- **URL:** `{{baseUrl}}/health`
- **Expected Status:** `200 OK`
- **Expected Response:**
```json
{
  "status": "UP"
}
```

---

### 2. Create Employee
- **Method:** `POST`
- **URL:** `{{baseUrl}}/api/v1/employees`
- **Headers:** `Content-Type: application/json`
- **Body (raw JSON):**
```json
{
  "first_name": "Arjun",
  "last_name": "Sharma",
  "email": "arjun.sharma@example.com",
  "phone": "+919876543210",
  "department": "Engineering",
  "designation": "Cloud DevOps Engineer",
  "salary": 85000.00,
  "joining_date": "2026-03-15"
}
```
- **Expected Status:** `201 Created`
- **Expected Response:**
```json
{
  "id": 1,
  "first_name": "Arjun",
  "last_name": "Sharma",
  "email": "arjun.sharma@example.com",
  "phone": "+919876543210",
  "department": "Engineering",
  "designation": "Cloud DevOps Engineer",
  "salary": 85000,
  "joining_date": "2026-03-15",
  "created_at": "2026-10-06T...",
  "updated_at": "2026-10-06T..."
}
```

---

### 3. List Employees
- **Method:** `GET`
- **URL:** `{{baseUrl}}/api/v1/employees?page=1&limit=10`
- **Expected Status:** `200 OK`
- **Expected Response:** Array containing the created employee records with pagination metadata.

---

### 4. Get Employee by ID
- **Method:** `GET`
- **URL:** `{{baseUrl}}/api/v1/employees/1`
- **Expected Status:** `200 OK`
- **Expected Response:** JSON object for employee ID 1.

---

### 5. Update Employee
- **Method:** `PUT`
- **URL:** `{{baseUrl}}/api/v1/employees/1`
- **Headers:** `Content-Type: application/json`
- **Body (raw JSON):**
```json
{
  "first_name": "Arjun",
  "last_name": "Sharma",
  "email": "arjun.sharma@example.com",
  "phone": "+919876543210",
  "department": "Infrastructure",
  "designation": "Senior Platform Engineer",
  "salary": 95000.00,
  "joining_date": "2026-03-15"
}
```
- **Expected Status:** `200 OK`
- **Expected Response:** Updated employee details showing new designation and salary.

---

### 6. Delete Employee
- **Method:** `DELETE`
- **URL:** `{{baseUrl}}/api/v1/employees/1`
- **Expected Status:** `204 No Content`

---

## 13. Troubleshooting Common Issues

### Issue 1: PVC Remains in `Pending` Status
- **Root Cause A:** `WaitForFirstConsumer` is waiting for MySQL pod. If MySQL pod is not deployed or cannot be scheduled, the PVC will stay pending. Check MySQL pod events:
  ```powershell
  kubectl describe pod -l app=employee-management-mysql
  ```
- **Root Cause B:** EBS CSI Driver is missing or inactive. Check:
  ```powershell
  kubectl get pods -n kube-system -l app.kubernetes.io/name=aws-ebs-csi-driver
  ```
- **Root Cause C:** EC2 IAM role lacks EBS permissions. Ensure `AmazonEBSCSIDriverPolicy` is attached to your worker node IAM role.
- **Inspect PVC Events:**
  ```powershell
  kubectl describe pvc mysql-pvc
  ```

---

### Issue 2: MySQL Pod In `CrashLoopBackOff`
- **Check Logs:**
  ```powershell
  kubectl logs -l app=employee-management-mysql
  ```
- **Common Fix:** Verify the secret `mysql-secret` exists and contains the key `root-password`. If the password contains special characters, ensure it was escaped properly when creating the secret.

---

### Issue 3: Migration Job Failed
- **Inspect Job Details:**
  ```powershell
  kubectl describe job employee-management-migration
  ```
- **View Migration Pod Logs:**
  ```powershell
  kubectl logs -l app=employee-management-migration
  ```
- **Recreate the Job:** If the job reached `backoffLimit` before MySQL was fully ready:
  ```powershell
  kubectl delete job employee-management-migration
  kubectl apply -f k8s/eks/migration.yaml
  ```

---

### Issue 4: Go Pods `CrashLoopBackOff` or Unhealthy Probes
- **Check Go Logs:**
  ```powershell
  kubectl logs -l app=employee-management-go
  ```
- If the logs show `dial tcp: lookup employee-management-mysql: no such host` or connection refused:
  - Check whether MySQL Service exists: `kubectl get svc employee-management-mysql`
  - Check whether CoreDNS is healthy: `kubectl get pods -n kube-system -l k8s-app=kube-dns`
  - Ensure the database migration has completed so the `employees` table exists.

---

### Issue 5: LoadBalancer `EXTERNAL-IP` Shows `<pending>`
- AWS Load Balancer creation typically takes 1 to 3 minutes.
- Check Service events:
  ```powershell
  kubectl describe svc employee-management-go
  ```
- If an authorization error is reported, verify that your EKS cluster security group and subnets permit external load balancers, and that public subnets are tagged:
  `kubernetes.io/role/elb = 1`

---

## 14. Resource Teardown & Cleanup Commands

To prevent unnecessary AWS cloud charges in your learning account, tear down resources in reverse order:

```powershell
# 1. Delete LoadBalancer Service first to trigger AWS Load Balancer deletion
kubectl delete -f k8s/eks/go-app-service.yaml

# 2. Delete Go Application Deployment
kubectl delete -f k8s/eks/go-app-deployment.yaml

# 3. Delete Migration Job and ConfigMap
kubectl delete -f k8s/eks/migration.yaml
kubectl delete -f k8s/eks/migration-configmap.yaml

# 4. Delete MySQL Service and Deployment
kubectl delete -f k8s/eks/mysql-service.yaml
kubectl delete -f k8s/eks/mysql-deployment.yaml

# 5. Delete PersistentVolumeClaim (triggers EBS volume deletion due to reclaimPolicy: Delete)
kubectl delete -f k8s/eks/mysql-pvc.yaml

# 6. Delete StorageClass, ConfigMap, and Secret
kubectl delete -f k8s/eks/storageclass.yaml
kubectl delete -f k8s/eks/configmap.yaml
kubectl delete secret mysql-secret
```

Verify everything is deleted:
```powershell
kubectl get all
kubectl get pvc
kubectl get pv
```
