# Employee Management Platform — API Design

## 1. Base URL

Local development:

```text
http://localhost:<port>/api/v1
```

The exact port is defined by runtime configuration.

## 2. Health Endpoint

```http
GET /health
```

Purpose:

- Confirm that the HTTP application process is responding.
- Provide a target for container/Kubernetes health checks.

The initial health endpoint should remain lightweight. Database dependency checks can be added separately if required.

## 3. Create Employee

```http
POST /api/v1/employees
Content-Type: application/json
```

Example request:

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

Response:

```http
201 Created
```

Example:

```json
{
  "id": 1,
  "first_name": "Rahul",
  "last_name": "Kumar",
  "email": "rahul.kumar@example.com",
  "phone": "+919876543210",
  "department": "Engineering",
  "designation": "Software Engineer",
  "salary": 65000,
  "joining_date": "2026-09-24",
  "created_at": "2026-09-24T10:00:00Z",
  "updated_at": "2026-09-24T10:00:00Z"
}
```

## 4. List Employees

```http
GET /api/v1/employees
```

Initial response:

```http
200 OK
```

Example:

```json
{
  "data": [
    {
      "id": 1,
      "first_name": "Rahul",
      "last_name": "Kumar",
      "email": "rahul.kumar@example.com",
      "phone": "+919876543210",
      "department": "Engineering",
      "designation": "Software Engineer",
      "salary": 65000,
      "joining_date": "2026-09-24",
      "created_at": "2026-09-24T10:00:00Z",
      "updated_at": "2026-09-24T10:00:00Z"
    }
  ]
}
```

Pagination can be introduced after the basic endpoint works.

## 5. Get Employee

```http
GET /api/v1/employees/{id}
```

Success:

```http
200 OK
```

Missing employee:

```http
404 Not Found
```

## 6. Update Employee

```http
PUT /api/v1/employees/{id}
Content-Type: application/json
```

The initial design treats PUT as a complete update of mutable employee fields.

Success:

```http
200 OK
```

Missing employee:

```http
404 Not Found
```

## 7. Delete Employee

```http
DELETE /api/v1/employees/{id}
```

Success:

```http
204 No Content
```

Missing employee:

```http
404 Not Found
```

## 8. Error Contract

Use a consistent structure:

```json
{
  "error": {
    "code": "EMPLOYEE_NOT_FOUND",
    "message": "Employee not found"
  }
}
```

Validation example:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid employee data",
    "details": {
      "email": "must be a valid email"
    }
  }
}
```

## 9. Error Codes

Initial codes:

- `VALIDATION_ERROR`
- `EMPLOYEE_NOT_FOUND`
- `EMPLOYEE_EMAIL_CONFLICT`
- `DATABASE_ERROR`
- `INTERNAL_SERVER_ERROR`

## 10. API Versioning

All business endpoints use:

```text
/api/v1
```

This creates a stable boundary for future incompatible API versions.

## 11. API Design Principles

- JSON request/response format.
- Consistent HTTP status codes.
- No raw database errors.
- Stable error schema.
- Versioned business endpoints.
- Health endpoint outside the business API version.
