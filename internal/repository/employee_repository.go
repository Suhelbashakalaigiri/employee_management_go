package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	apperrors "employee-management-platform/internal/errors"
	"employee-management-platform/internal/model"
	"github.com/go-sql-driver/mysql"
)

const (
	mysqlErrDuplicateEntry = 1062
)

// EmployeeRepository defines the persistence operations for employees.
type EmployeeRepository interface {
	Create(ctx context.Context, emp *model.Employee) error
	GetByID(ctx context.Context, id uint64) (*model.Employee, error)
	GetAll(ctx context.Context) ([]model.Employee, error)
	Update(ctx context.Context, emp *model.Employee) error
	Delete(ctx context.Context, id uint64) error
	GetByEmail(ctx context.Context, email string) (*model.Employee, error)
}

// mysqlEmployeeRepository implements EmployeeRepository using standard database/sql.
type mysqlEmployeeRepository struct {
	db *sql.DB
}

// NewEmployeeRepository returns a new instance of mysqlEmployeeRepository.
func NewEmployeeRepository(db *sql.DB) EmployeeRepository {
	return &mysqlEmployeeRepository{db: db}
}

// Create inserts a new employee record and populates the generated ID and timestamps.
func (r *mysqlEmployeeRepository) Create(ctx context.Context, emp *model.Employee) error {
	query := `
		INSERT INTO employees (
			first_name, last_name, email, phone, department, designation, salary, joining_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	valDate, err := emp.JoiningDate.Value()
	if err != nil {
		return fmt.Errorf("invalid joining date: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		emp.FirstName,
		emp.LastName,
		emp.Email,
		emp.Phone,
		emp.Department,
		emp.Designation,
		emp.Salary,
		valDate,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
			return apperrors.ErrDuplicateEmail
		}
		return fmt.Errorf("repository: failed to insert employee: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: failed to retrieve last insert id: %w", err)
	}
	emp.ID = uint64(id)

	// Fetch database-generated timestamps to ensure absolute accuracy
	created, err := r.GetByID(ctx, emp.ID)
	if err != nil {
		return fmt.Errorf("repository: failed to reload created employee: %w", err)
	}
	emp.CreatedAt = created.CreatedAt
	emp.UpdatedAt = created.UpdatedAt

	return nil
}

// GetByID retrieves a single employee by primary key.
func (r *mysqlEmployeeRepository) GetByID(ctx context.Context, id uint64) (*model.Employee, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at
		FROM employees
		WHERE id = ?
	`

	var emp model.Employee
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&emp.ID,
		&emp.FirstName,
		&emp.LastName,
		&emp.Email,
		&emp.Phone,
		&emp.Department,
		&emp.Designation,
		&emp.Salary,
		&emp.JoiningDate,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("repository: failed to query employee by id %d: %w", id, err)
	}

	return &emp, nil
}

// GetAll retrieves all employees ordered by primary key ascending.
func (r *mysqlEmployeeRepository) GetAll(ctx context.Context) ([]model.Employee, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at
		FROM employees
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to query all employees: %w", err)
	}
	defer rows.Close()

	employees := make([]model.Employee, 0)
	for rows.Next() {
		var emp model.Employee
		if err := rows.Scan(
			&emp.ID,
			&emp.FirstName,
			&emp.LastName,
			&emp.Email,
			&emp.Phone,
			&emp.Department,
			&emp.Designation,
			&emp.Salary,
			&emp.JoiningDate,
			&emp.CreatedAt,
			&emp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: failed to scan employee row: %w", err)
		}
		employees = append(employees, emp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: error iterating employee rows: %w", err)
	}

	return employees, nil
}

// Update updates mutable fields of an existing employee.
func (r *mysqlEmployeeRepository) Update(ctx context.Context, emp *model.Employee) error {
	query := `
		UPDATE employees
		SET first_name = ?, last_name = ?, email = ?, phone = ?, department = ?, designation = ?, salary = ?, joining_date = ?
		WHERE id = ?
	`

	valDate, err := emp.JoiningDate.Value()
	if err != nil {
		return fmt.Errorf("invalid joining date: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		emp.FirstName,
		emp.LastName,
		emp.Email,
		emp.Phone,
		emp.Department,
		emp.Designation,
		emp.Salary,
		valDate,
		emp.ID,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
			return apperrors.ErrDuplicateEmail
		}
		return fmt.Errorf("repository: failed to update employee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Verify if the employee exists or if fields simply remained unchanged
		_, err := r.GetByID(ctx, emp.ID)
		if err != nil {
			return err
		}
	}

	// Fetch updated timestamps
	updated, err := r.GetByID(ctx, emp.ID)
	if err != nil {
		return fmt.Errorf("repository: failed to reload updated employee: %w", err)
	}
	emp.CreatedAt = updated.CreatedAt
	emp.UpdatedAt = updated.UpdatedAt

	return nil
}

// Delete removes an employee by primary key.
func (r *mysqlEmployeeRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM employees WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: failed to delete employee id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

// GetByEmail retrieves an employee by email address.
func (r *mysqlEmployeeRepository) GetByEmail(ctx context.Context, email string) (*model.Employee, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at
		FROM employees
		WHERE email = ?
	`

	var emp model.Employee
	row := r.db.QueryRowContext(ctx, query, email)
	err := row.Scan(
		&emp.ID,
		&emp.FirstName,
		&emp.LastName,
		&emp.Email,
		&emp.Phone,
		&emp.Department,
		&emp.Designation,
		&emp.Salary,
		&emp.JoiningDate,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("repository: failed to query employee by email: %w", err)
	}

	return &emp, nil
}
