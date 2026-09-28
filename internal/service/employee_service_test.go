package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "employee-management-platform/internal/errors"
	"employee-management-platform/internal/model"
	"employee-management-platform/internal/service"
	"github.com/shopspring/decimal"
)

type mockRepository struct {
	employees     map[uint64]*model.Employee
	nextID        uint64
	createErr     error
	getErr        error
	getAllErr     error
	updateErr     error
	deleteErr     error
	getByEmailErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		employees: make(map[uint64]*model.Employee),
		nextID:    1,
	}
}

func (m *mockRepository) Create(ctx context.Context, emp *model.Employee) error {
	if m.createErr != nil {
		return m.createErr
	}
	emp.ID = m.nextID
	m.nextID++
	now := time.Now().UTC()
	emp.CreatedAt = now
	emp.UpdatedAt = now

	stored := *emp
	m.employees[emp.ID] = &stored
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uint64) (*model.Employee, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	emp, ok := m.employees[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	res := *emp
	return &res, nil
}

func (m *mockRepository) GetAll(ctx context.Context) ([]model.Employee, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	list := make([]model.Employee, 0, len(m.employees))
	for _, emp := range m.employees {
		list = append(list, *emp)
	}
	return list, nil
}

func (m *mockRepository) Update(ctx context.Context, emp *model.Employee) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.employees[emp.ID]; !ok {
		return apperrors.ErrNotFound
	}
	emp.UpdatedAt = time.Now().UTC()
	stored := *emp
	m.employees[emp.ID] = &stored
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uint64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.employees[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.employees, id)
	return nil
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*model.Employee, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	for _, emp := range m.employees {
		if emp.Email == email {
			res := *emp
			return &res, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func TestServiceCreateEmployeeSuccess(t *testing.T) {
	repo := newMockRepository()
	svc := service.NewEmployeeService(repo)

	req := &model.CreateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul.kumar@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	}

	emp, err := svc.CreateEmployee(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if emp.ID != 1 {
		t.Errorf("expected ID 1, got %d", emp.ID)
	}
	if emp.Email != "rahul.kumar@example.com" {
		t.Errorf("expected rahul.kumar@example.com, got %s", emp.Email)
	}
}

func TestServiceCreateEmployeeValidationError(t *testing.T) {
	repo := newMockRepository()
	svc := service.NewEmployeeService(repo)

	// Missing required fields
	req := &model.CreateEmployeeRequest{
		FirstName: "",
	}

	_, err := svc.CreateEmployee(context.Background(), req)
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %v", err)
	}

	if appErr.Code != apperrors.CodeValidationError {
		t.Errorf("expected code %s, got %s", apperrors.CodeValidationError, appErr.Code)
	}
}

func TestServiceCreateEmployeeDuplicateEmail(t *testing.T) {
	repo := newMockRepository()
	svc := service.NewEmployeeService(repo)

	req := &model.CreateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "duplicate@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	}

	// First creation succeeds
	_, err := svc.CreateEmployee(context.Background(), req)
	if err != nil {
		t.Fatalf("expected first creation to succeed, got: %v", err)
	}

	// Second creation with same email fails with conflict
	_, err = svc.CreateEmployee(context.Background(), req)
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %v", err)
	}

	if appErr.Code != apperrors.CodeEmployeeEmailConflict {
		t.Errorf("expected code %s, got %s", apperrors.CodeEmployeeEmailConflict, appErr.Code)
	}
}

func TestServiceGetEmployeeByID(t *testing.T) {
	repo := newMockRepository()
	svc := service.NewEmployeeService(repo)

	// Not found
	_, err := svc.GetEmployeeByID(context.Background(), 100)
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeEmployeeNotFound {
		t.Errorf("expected not found error, got: %v", err)
	}

	// Create and get
	created, _ := svc.CreateEmployee(context.Background(), &model.CreateEmployeeRequest{
		FirstName:   "Priya",
		LastName:    "Sharma",
		Email:       "priya@example.com",
		Phone:       "+919876543211",
		Department:  "Product",
		Designation: "Product Manager",
		Salary:      decimal.NewFromInt(75000),
		JoiningDate: model.Date("2026-09-25"),
	})

	emp, err := svc.GetEmployeeByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if emp.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, emp.ID)
	}
}

func TestServiceUpdateEmployee(t *testing.T) {
	repo := newMockRepository()
	svc := service.NewEmployeeService(repo)

	created, _ := svc.CreateEmployee(context.Background(), &model.CreateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	})

	// Update existing
	updateReq := &model.UpdateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Verma",
		Email:       "rahul@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Senior Software Engineer",
		Salary:      decimal.NewFromInt(80000),
		JoiningDate: model.Date("2026-09-24"),
	}

	updated, err := svc.UpdateEmployee(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Fatalf("expected update to succeed, got: %v", err)
	}
	if updated.Designation != "Senior Software Engineer" || updated.LastName != "Verma" {
		t.Errorf("expected updated fields, got %+v", updated)
	}

	// Update not found
	_, err = svc.UpdateEmployee(context.Background(), 999, updateReq)
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeEmployeeNotFound {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestServiceDeleteEmployee(t *testing.T) {
	repo := newMockRepository()
	svc := service.NewEmployeeService(repo)

	created, _ := svc.CreateEmployee(context.Background(), &model.CreateEmployeeRequest{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	})

	// Delete
	err := svc.DeleteEmployee(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected delete to succeed, got: %v", err)
	}

	// Verify it is gone
	_, err = svc.GetEmployeeByID(context.Background(), created.ID)
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeEmployeeNotFound {
		t.Errorf("expected not found after deletion, got: %v", err)
	}

	// Delete again should return not found
	err = svc.DeleteEmployee(context.Background(), created.ID)
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeEmployeeNotFound {
		t.Errorf("expected not found on second deletion, got: %v", err)
	}
}
