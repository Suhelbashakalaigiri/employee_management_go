package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "employee-management-platform/internal/errors"
	"employee-management-platform/internal/handler"
	"employee-management-platform/internal/model"
	"employee-management-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// inMemoryRepo provides a lightweight in-memory storage for handler integration testing.
type inMemoryRepo struct {
	employees map[uint64]*model.Employee
	nextID    uint64
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{
		employees: make(map[uint64]*model.Employee),
		nextID:    1,
	}
}

func (r *inMemoryRepo) Create(ctx context.Context, emp *model.Employee) error {
	for _, e := range r.employees {
		if e.Email == emp.Email {
			return apperrors.ErrDuplicateEmail
		}
	}
	emp.ID = r.nextID
	r.nextID++
	now := time.Now().UTC()
	emp.CreatedAt = now
	emp.UpdatedAt = now
	stored := *emp
	r.employees[emp.ID] = &stored
	return nil
}

func (r *inMemoryRepo) GetByID(ctx context.Context, id uint64) (*model.Employee, error) {
	emp, ok := r.employees[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	res := *emp
	return &res, nil
}

func (r *inMemoryRepo) GetAll(ctx context.Context) ([]model.Employee, error) {
	list := make([]model.Employee, 0, len(r.employees))
	for _, emp := range r.employees {
		list = append(list, *emp)
	}
	return list, nil
}

func (r *inMemoryRepo) Update(ctx context.Context, emp *model.Employee) error {
	existing, ok := r.employees[emp.ID]
	if !ok {
		return apperrors.ErrNotFound
	}
	for _, e := range r.employees {
		if e.ID != emp.ID && e.Email == emp.Email {
			return apperrors.ErrDuplicateEmail
		}
	}
	emp.CreatedAt = existing.CreatedAt
	emp.UpdatedAt = time.Now().UTC()
	stored := *emp
	r.employees[emp.ID] = &stored
	return nil
}

func (r *inMemoryRepo) Delete(ctx context.Context, id uint64) error {
	if _, ok := r.employees[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(r.employees, id)
	return nil
}

func (r *inMemoryRepo) GetByEmail(ctx context.Context, email string) (*model.Employee, error) {
	for _, emp := range r.employees {
		if emp.Email == email {
			res := *emp
			return &res, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func setupTestRouter() (*gin.Engine, *inMemoryRepo) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := newInMemoryRepo()
	svc := service.NewEmployeeService(repo)
	h := handler.NewEmployeeHandler(svc)

	v1 := r.Group("/api/v1")
	{
		employees := v1.Group("/employees")
		{
			employees.POST("", h.CreateEmployee)
			employees.GET("", h.GetAllEmployees)
			employees.GET("/:id", h.GetEmployeeByID)
			employees.PUT("/:id", h.UpdateEmployee)
			employees.DELETE("/:id", h.DeleteEmployee)
		}
	}

	return r, repo
}

func TestHandlerCreateEmployeeSuccess(t *testing.T) {
	r, _ := setupTestRouter()

	reqBody := model.CreateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul.kumar@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var res model.Employee
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.ID == 0 || res.FirstName != "Rahul" || res.Email != "rahul.kumar@example.com" {
		t.Errorf("unexpected created employee response: %+v", res)
	}
}

func TestHandlerCreateEmployeeInvalidRequest(t *testing.T) {
	r, _ := setupTestRouter()

	// Missing required fields
	reqBody := model.CreateEmployeeRequest{
		FirstName: "",
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d: %s", w.Code, w.Body.String())
	}

	var errRes apperrors.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errRes); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errRes.Error.Code != apperrors.CodeValidationError {
		t.Errorf("expected error code %s, got %s", apperrors.CodeValidationError, errRes.Error.Code)
	}
	if len(errRes.Error.Details) == 0 {
		t.Errorf("expected error details, got empty")
	}
}

func TestHandlerCreateEmployeeDuplicateEmail(t *testing.T) {
	r, _ := setupTestRouter()

	reqBody := model.CreateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "duplicate@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	}

	jsonBytes, _ := json.Marshal(reqBody)

	// First request succeeds
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewBuffer(jsonBytes))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created on first request, got %d", w1.Code)
	}

	// Second request with duplicate email must fail with 409 Conflict
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewBuffer(jsonBytes))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected status 409 Conflict, got %d: %s", w2.Code, w2.Body.String())
	}

	var errRes apperrors.ErrorResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &errRes); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}

	if errRes.Error.Code != apperrors.CodeEmployeeEmailConflict {
		t.Errorf("expected code %s, got %s", apperrors.CodeEmployeeEmailConflict, errRes.Error.Code)
	}
}

func TestHandlerGetEmployeeByIDSuccess(t *testing.T) {
	r, repo := setupTestRouter()

	_ = repo.Create(context.Background(), &model.Employee{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul.kumar@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/employees/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res model.Employee
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse employee: %v", err)
	}

	if res.ID != 1 || res.FirstName != "Rahul" {
		t.Errorf("unexpected employee returned: %+v", res)
	}
}

func TestHandlerGetEmployeeByIDNotFound(t *testing.T) {
	r, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/employees/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d: %s", w.Code, w.Body.String())
	}

	var errRes apperrors.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errRes); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}

	if errRes.Error.Code != apperrors.CodeEmployeeNotFound {
		t.Errorf("expected error code %s, got %s", apperrors.CodeEmployeeNotFound, errRes.Error.Code)
	}
}

func TestHandlerListEmployees(t *testing.T) {
	r, repo := setupTestRouter()

	_ = repo.Create(context.Background(), &model.Employee{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	})
	_ = repo.Create(context.Background(), &model.Employee{
		FirstName:   "Priya",
		LastName:    "Sharma",
		Email:       "priya@example.com",
		Phone:       "+919876543211",
		Department:  "Product",
		Designation: "Product Manager",
		Salary:      decimal.NewFromInt(75000),
		JoiningDate: model.Date("2026-09-25"),
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/employees", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res model.EmployeeListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if len(res.Data) != 2 {
		t.Fatalf("expected 2 employees in list, got %d", len(res.Data))
	}
}

func TestHandlerUpdateEmployee(t *testing.T) {
	r, repo := setupTestRouter()

	_ = repo.Create(context.Background(), &model.Employee{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	})

	updateBody := model.UpdateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Verma",
		Email:       "rahul@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Lead Engineer",
		Salary:      decimal.NewFromInt(85000),
		JoiningDate: model.Date("2026-09-24"),
	}

	jsonBytes, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/employees/1", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res model.Employee
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse updated employee: %v", err)
	}

	if res.LastName != "Verma" || res.Designation != "Lead Engineer" {
		t.Errorf("fields were not properly updated: %+v", res)
	}

	// Test 404 for updating non-existent employee
	reqNotFound, _ := http.NewRequest(http.MethodPut, "/api/v1/employees/999", bytes.NewBuffer(jsonBytes))
	reqNotFound.Header.Set("Content-Type", "application/json")
	wNotFound := httptest.NewRecorder()
	r.ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", wNotFound.Code)
	}
}

func TestHandlerDeleteEmployee(t *testing.T) {
	r, repo := setupTestRouter()

	_ = repo.Create(context.Background(), &model.Employee{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	})

	// Delete existing employee -> 204 No Content
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/employees/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d", w.Code)
	}

	// Delete non-existent employee -> 404 Not Found
	reqNotFound, _ := http.NewRequest(http.MethodDelete, "/api/v1/employees/1", nil)
	wNotFound := httptest.NewRecorder()
	r.ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d", wNotFound.Code)
	}

	var errRes apperrors.ErrorResponse
	if err := json.Unmarshal(wNotFound.Body.Bytes(), &errRes); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if errRes.Error.Code != apperrors.CodeEmployeeNotFound {
		t.Errorf("expected %s, got %s", apperrors.CodeEmployeeNotFound, errRes.Error.Code)
	}
}

func TestHandlerInvalidPathID(t *testing.T) {
	r, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/employees/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request for non-integer ID, got %d: %s", w.Code, w.Body.String())
	}
}
