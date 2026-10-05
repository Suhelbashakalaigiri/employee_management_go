# Employee Management Platform — Database Design

## 1. Database

Development and initial Kubernetes environments use:

```text
MySQL
```

The AWS target environment uses:

```text
Amazon RDS for MySQL
```

## 2. Database Name

Suggested database:

```text
employee_management
```

The exact name should be configurable through environment variables.

## 3. Employee Table

```sql
CREATE TABLE employees (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(30) NOT NULL,
    department VARCHAR(100) NOT NULL,
    designation VARCHAR(100) NOT NULL,
    salary DECIMAL(12,2) NOT NULL,
    joining_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uq_employees_email UNIQUE (email)
);
```

The final migration syntax will be reviewed against the selected migration/tooling approach before implementation.

## 4. Constraints

- `id` is the primary key.
- `email` is unique.
- Required fields use `NOT NULL`.
- Salary uses a fixed-point decimal type rather than floating-point storage.
- Timestamps are maintained by the database/application strategy selected during implementation.

## 5. Indexing

Initial index:

```text
UNIQUE(email)
```

Additional indexes should only be introduced when supported by actual query requirements.

Potential future indexes:

```text
department
joining_date
```

These should not be added blindly.

## 6. Migrations

Schema changes must be version-controlled.

Example:

```text
migrations/
├── 000001_create_employees.up.sql
└── 000001_create_employees.down.sql
```

A migration tool may be selected during implementation. The tool must support deterministic, repeatable migration execution.

## 7. Database Access

The Go application should use a connection pool.

Configuration should include:

```text
DB_HOST
DB_PORT
DB_NAME
DB_USER
DB_PASSWORD
```

Connection limits and timeouts should be configurable.

## 8. Transactions

Transactions should be used where multiple database changes must succeed or fail together.

Basic single-row CRUD operations do not automatically require explicit transactions if one atomic SQL statement is sufficient.

## 9. Local Database

Docker Compose should provide a MySQL container for reproducible local development.

```text
employee-api
     |
     | employee-network
     v
mysql
```

The application must connect to the Compose service name rather than assuming `localhost` from inside the container.

## 10. AWS Database

The production-style AWS environment will use:

```text
EKS
 |
 +-- Employee API
 |
 v
RDS MySQL
```

RDS remains outside the Kubernetes cluster.

## 11. Database Backup and HA

These concerns are part of the later AWS production-hardening stage. The initial project focuses on correct connectivity and deployment rather than full database operations.
