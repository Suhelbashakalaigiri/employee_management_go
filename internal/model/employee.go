package model

import (
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Standard phone format regex (allows optional leading '+', digits, dashes, spaces, parentheses, length 7 to 30)
var phoneRegex = regexp.MustCompile(`^\+?[0-9\s\-()]{7,30}$`)

// Employee represents an employee entity in the domain and database.
type Employee struct {
	ID          uint64          `json:"id"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Department  string          `json:"department"`
	Designation string          `json:"designation"`
	Salary      decimal.Decimal `json:"salary"`
	JoiningDate Date            `json:"joining_date"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CreateEmployeeRequest represents the incoming JSON payload for creating an employee.
type CreateEmployeeRequest struct {
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Department  string          `json:"department"`
	Designation string          `json:"designation"`
	Salary      decimal.Decimal `json:"salary"`
	JoiningDate Date            `json:"joining_date"`
}

// UpdateEmployeeRequest represents the incoming JSON payload for updating an employee.
type UpdateEmployeeRequest struct {
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Department  string          `json:"department"`
	Designation string          `json:"designation"`
	Salary      decimal.Decimal `json:"salary"`
	JoiningDate Date            `json:"joining_date"`
}

// EmployeeListResponse represents the JSON response wrapper for a list of employees.
type EmployeeListResponse struct {
	Data []Employee `json:"data"`
}

// ValidateCreateRequest performs business validation on CreateEmployeeRequest.
// It returns a map of field name -> error message, which is empty if valid.
func (r *CreateEmployeeRequest) Validate() map[string]string {
	details := make(map[string]string)

	if strings.TrimSpace(r.FirstName) == "" {
		details["first_name"] = "first_name is required"
	} else if len(r.FirstName) > 100 {
		details["first_name"] = "first_name cannot exceed 100 characters"
	}

	if strings.TrimSpace(r.LastName) == "" {
		details["last_name"] = "last_name is required"
	} else if len(r.LastName) > 100 {
		details["last_name"] = "last_name cannot exceed 100 characters"
	}

	cleanEmail := strings.TrimSpace(r.Email)
	if cleanEmail == "" {
		details["email"] = "email is required"
	} else if len(cleanEmail) > 255 {
		details["email"] = "email cannot exceed 255 characters"
	} else if _, err := mail.ParseAddress(cleanEmail); err != nil || !strings.Contains(cleanEmail, ".") {
		details["email"] = "must be a valid email address"
	}

	cleanPhone := strings.TrimSpace(r.Phone)
	if cleanPhone == "" {
		details["phone"] = "phone is required"
	} else if !phoneRegex.MatchString(cleanPhone) {
		details["phone"] = "phone must be a valid phone number (7-30 characters, digits and optional +, -, ())"
	}

	if strings.TrimSpace(r.Department) == "" {
		details["department"] = "department is required"
	} else if len(r.Department) > 100 {
		details["department"] = "department cannot exceed 100 characters"
	}

	if strings.TrimSpace(r.Designation) == "" {
		details["designation"] = "designation is required"
	} else if len(r.Designation) > 100 {
		details["designation"] = "designation cannot exceed 100 characters"
	}

	if r.Salary.IsNegative() {
		details["salary"] = "salary must be non-negative"
	}

	if !r.JoiningDate.IsValid() {
		details["joining_date"] = "joining_date is required and must be in YYYY-MM-DD format"
	}

	return details
}

// ValidateUpdateRequest performs business validation on UpdateEmployeeRequest.
func (r *UpdateEmployeeRequest) Validate() map[string]string {
	details := make(map[string]string)

	if strings.TrimSpace(r.FirstName) == "" {
		details["first_name"] = "first_name is required"
	} else if len(r.FirstName) > 100 {
		details["first_name"] = "first_name cannot exceed 100 characters"
	}

	if strings.TrimSpace(r.LastName) == "" {
		details["last_name"] = "last_name is required"
	} else if len(r.LastName) > 100 {
		details["last_name"] = "last_name cannot exceed 100 characters"
	}

	cleanEmail := strings.TrimSpace(r.Email)
	if cleanEmail == "" {
		details["email"] = "email is required"
	} else if len(cleanEmail) > 255 {
		details["email"] = "email cannot exceed 255 characters"
	} else if _, err := mail.ParseAddress(cleanEmail); err != nil || !strings.Contains(cleanEmail, ".") {
		details["email"] = "must be a valid email address"
	}

	cleanPhone := strings.TrimSpace(r.Phone)
	if cleanPhone == "" {
		details["phone"] = "phone is required"
	} else if !phoneRegex.MatchString(cleanPhone) {
		details["phone"] = "phone must be a valid phone number (7-30 characters, digits and optional +, -, ())"
	}

	if strings.TrimSpace(r.Department) == "" {
		details["department"] = "department is required"
	} else if len(r.Department) > 100 {
		details["department"] = "department cannot exceed 100 characters"
	}

	if strings.TrimSpace(r.Designation) == "" {
		details["designation"] = "designation is required"
	} else if len(r.Designation) > 100 {
		details["designation"] = "designation cannot exceed 100 characters"
	}

	if r.Salary.IsNegative() {
		details["salary"] = "salary must be non-negative"
	}

	if !r.JoiningDate.IsValid() {
		details["joining_date"] = "joining_date is required and must be in YYYY-MM-DD format"
	}

	return details
}
