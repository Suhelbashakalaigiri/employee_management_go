package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"employee-management-platform/internal/model"
	"github.com/shopspring/decimal"
)

func TestDateJSONMarshaling(t *testing.T) {
	d := model.Date("2026-09-24")

	bytes, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("failed to marshal Date: %v", err)
	}

	if string(bytes) != `"2026-09-24"` {
		t.Errorf("expected %q, got %s", `"2026-09-24"`, string(bytes))
	}

	var unmarshaled model.Date
	if err := json.Unmarshal(bytes, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal Date: %v", err)
	}

	if unmarshaled != d {
		t.Errorf("expected %s, got %s", d, unmarshaled)
	}
}

func TestDateScan(t *testing.T) {
	var d model.Date

	// Scan time.Time
	now := time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC)
	if err := d.Scan(now); err != nil {
		t.Fatalf("failed to scan time.Time: %v", err)
	}
	if d != "2026-09-24" {
		t.Errorf("expected 2026-09-24, got %s", d)
	}

	// Scan []byte
	if err := d.Scan([]byte("2026-09-25 00:00:00")); err != nil {
		t.Fatalf("failed to scan []byte: %v", err)
	}
	if d != "2026-09-25" {
		t.Errorf("expected 2026-09-25, got %s", d)
	}

	// Scan string
	if err := d.Scan("2026-09-26"); err != nil {
		t.Fatalf("failed to scan string: %v", err)
	}
	if d != "2026-09-26" {
		t.Errorf("expected 2026-09-26, got %s", d)
	}
}

func TestCreateEmployeeValidation(t *testing.T) {
	validReq := model.CreateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul.kumar@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	}

	if details := validReq.Validate(); len(details) != 0 {
		t.Errorf("expected no validation errors for valid request, got: %v", details)
	}

	// Test missing required fields
	emptyReq := model.CreateEmployeeRequest{}
	details := emptyReq.Validate()

	requiredFields := []string{"first_name", "last_name", "email", "phone", "department", "designation", "joining_date"}
	for _, field := range requiredFields {
		if _, ok := details[field]; !ok {
			t.Errorf("expected validation error for field %q, but got none", field)
		}
	}

	// Test invalid email
	invalidEmailReq := validReq
	invalidEmailReq.Email = "not-an-email"
	if d := invalidEmailReq.Validate(); d["email"] == "" {
		t.Errorf("expected validation error for invalid email")
	}

	// Test negative salary
	negSalaryReq := validReq
	negSalaryReq.Salary = decimal.NewFromFloat(-100.50)
	if d := negSalaryReq.Validate(); d["salary"] == "" {
		t.Errorf("expected validation error for negative salary")
	}

	// Test invalid date format
	invalidDateReq := validReq
	invalidDateReq.JoiningDate = "24-09-2026"
	if d := invalidDateReq.Validate(); d["joining_date"] == "" {
		t.Errorf("expected validation error for invalid joining_date")
	}
}
