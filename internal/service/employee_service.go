package service

import (
	"context"
	"errors"
	"strings"

	apperrors "employee-management-platform/internal/errors"
	"employee-management-platform/internal/model"
	"employee-management-platform/internal/repository"
)

// EmployeeService defines the business operations for employee management.
type EmployeeService interface {
	CreateEmployee(ctx context.Context, req *model.CreateEmployeeRequest) (*model.Employee, error)
	GetEmployeeByID(ctx context.Context, id uint64) (*model.Employee, error)
	GetAllEmployees(ctx context.Context) ([]model.Employee, error)
	UpdateEmployee(ctx context.Context, id uint64, req *model.UpdateEmployeeRequest) (*model.Employee, error)
	DeleteEmployee(ctx context.Context, id uint64) error
}

type employeeService struct {
	repo repository.EmployeeRepository
}

// NewEmployeeService constructs an EmployeeService with the given repository.
func NewEmployeeService(repo repository.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

// CreateEmployee validates the input, checks email uniqueness, and creates a new employee.
func (s *employeeService) CreateEmployee(ctx context.Context, req *model.CreateEmployeeRequest) (*model.Employee, error) {
	if validationDetails := req.Validate(); len(validationDetails) > 0 {
		return nil, apperrors.NewValidationError("Invalid employee data", validationDetails)
	}

	cleanEmail := strings.TrimSpace(req.Email)

	// Business rule: proactive duplicate email check
	existing, err := s.repo.GetByEmail(ctx, cleanEmail)
	if err == nil && existing != nil {
		return nil, apperrors.NewConflictError("An employee with this email already exists")
	} else if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.NewDatabaseError(err)
	}

	emp := &model.Employee{
		FirstName:   strings.TrimSpace(req.FirstName),
		LastName:    strings.TrimSpace(req.LastName),
		Email:       cleanEmail,
		Phone:       strings.TrimSpace(req.Phone),
		Department:  strings.TrimSpace(req.Department),
		Designation: strings.TrimSpace(req.Designation),
		Salary:      req.Salary,
		JoiningDate: req.JoiningDate,
	}

	if err := s.repo.Create(ctx, emp); err != nil {
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			return nil, apperrors.NewConflictError("An employee with this email already exists")
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	return emp, nil
}

// GetEmployeeByID fetches an employee by ID or returns a not-found error.
func (s *employeeService) GetEmployeeByID(ctx context.Context, id uint64) (*model.Employee, error) {
	if id == 0 {
		return nil, apperrors.NewNotFoundError("Employee not found")
	}

	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.NewNotFoundError("Employee not found")
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	return emp, nil
}

// GetAllEmployees fetches all employees or returns an empty list.
func (s *employeeService) GetAllEmployees(ctx context.Context) ([]model.Employee, error) {
	employees, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	if employees == nil {
		employees = make([]model.Employee, 0)
	}

	return employees, nil
}

// UpdateEmployee validates input, ensures existence and uniqueness, and performs the update.
func (s *employeeService) UpdateEmployee(ctx context.Context, id uint64, req *model.UpdateEmployeeRequest) (*model.Employee, error) {
	if id == 0 {
		return nil, apperrors.NewNotFoundError("Employee not found")
	}

	if validationDetails := req.Validate(); len(validationDetails) > 0 {
		return nil, apperrors.NewValidationError("Invalid employee data", validationDetails)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.NewNotFoundError("Employee not found")
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	cleanEmail := strings.TrimSpace(req.Email)

	// If email is changing, verify that the new email does not collide with another employee
	if !strings.EqualFold(existing.Email, cleanEmail) {
		byEmail, err := s.repo.GetByEmail(ctx, cleanEmail)
		if err == nil && byEmail != nil && byEmail.ID != id {
			return nil, apperrors.NewConflictError("An employee with this email already exists")
		} else if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.NewDatabaseError(err)
		}
	}

	emp := &model.Employee{
		ID:          id,
		FirstName:   strings.TrimSpace(req.FirstName),
		LastName:    strings.TrimSpace(req.LastName),
		Email:       cleanEmail,
		Phone:       strings.TrimSpace(req.Phone),
		Department:  strings.TrimSpace(req.Department),
		Designation: strings.TrimSpace(req.Designation),
		Salary:      req.Salary,
		JoiningDate: req.JoiningDate,
	}

	if err := s.repo.Update(ctx, emp); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.NewNotFoundError("Employee not found")
		}
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			return nil, apperrors.NewConflictError("An employee with this email already exists")
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	return emp, nil
}

// DeleteEmployee removes an employee or returns a not-found error.
func (s *employeeService) DeleteEmployee(ctx context.Context, id uint64) error {
	if id == 0 {
		return apperrors.NewNotFoundError("Employee not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.NewNotFoundError("Employee not found")
		}
		return apperrors.NewDatabaseError(err)
	}

	return nil
}
