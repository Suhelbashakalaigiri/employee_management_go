package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	apperrors "employee-management-platform/internal/errors"
	"employee-management-platform/internal/model"
	"employee-management-platform/internal/repository"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
)

func setupMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.EmployeeRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	repo := repository.NewEmployeeRepository(db)
	return db, mock, repo
}

func TestRepositoryCreate(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	emp := &model.Employee{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul.kumar@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	}

	insertQuery := `INSERT INTO employees`
	mock.ExpectExec(insertQuery).
		WithArgs(
			emp.FirstName,
			emp.LastName,
			emp.Email,
			emp.Phone,
			emp.Department,
			emp.Designation,
			emp.Salary,
			"2026-09-24",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Reload query expectation
	selectQuery := `SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at FROM employees WHERE id = ?`
	rows := sqlmock.NewRows([]string{
		"id", "first_name", "last_name", "email", "phone", "department", "designation", "salary", "joining_date", "created_at", "updated_at",
	}).AddRow(1, emp.FirstName, emp.LastName, emp.Email, emp.Phone, emp.Department, emp.Designation, "65000.00", "2026-09-24", now, now)

	mock.ExpectQuery(selectQuery).WithArgs(uint64(1)).WillReturnRows(rows)

	err := repo.Create(ctx, emp)
	if err != nil {
		t.Fatalf("unexpected error creating employee: %v", err)
	}

	if emp.ID != 1 {
		t.Errorf("expected emp.ID to be 1, got %d", emp.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sql expectations: %v", err)
	}
}

func TestRepositoryCreateDuplicateEmail(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	ctx := context.Background()
	emp := &model.Employee{
		FirstName:   "Rahul",
		LastName:    "Kumar",
		Email:       "rahul.kumar@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Software Engineer",
		Salary:      decimal.NewFromInt(65000),
		JoiningDate: model.Date("2026-09-24"),
	}

	insertQuery := `INSERT INTO employees`
	mysqlErr := &mysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry 'rahul.kumar@example.com' for key 'uq_employees_email'",
	}
	mock.ExpectExec(insertQuery).WillReturnError(mysqlErr)

	err := repo.Create(ctx, emp)
	if !errors.Is(err, apperrors.ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got: %v", err)
	}
}

func TestRepositoryGetByID(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	selectQuery := `SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at FROM employees WHERE id = ?`
	rows := sqlmock.NewRows([]string{
		"id", "first_name", "last_name", "email", "phone", "department", "designation", "salary", "joining_date", "created_at", "updated_at",
	}).AddRow(1, "Rahul", "Kumar", "rahul.kumar@example.com", "+919876543210", "Engineering", "Software Engineer", "65000.00", "2026-09-24", now, now)

	mock.ExpectQuery(selectQuery).WithArgs(uint64(1)).WillReturnRows(rows)

	emp, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if emp.ID != 1 || emp.FirstName != "Rahul" || emp.Email != "rahul.kumar@example.com" {
		t.Errorf("unexpected employee data retrieved: %+v", emp)
	}

	// Test Not Found
	mock.ExpectQuery(selectQuery).WithArgs(uint64(999)).WillReturnError(sql.ErrNoRows)
	_, err = repo.GetByID(ctx, 999)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestRepositoryGetAll(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	selectQuery := `SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at FROM employees`
	rows := sqlmock.NewRows([]string{
		"id", "first_name", "last_name", "email", "phone", "department", "designation", "salary", "joining_date", "created_at", "updated_at",
	}).
		AddRow(1, "Rahul", "Kumar", "rahul.kumar@example.com", "+919876543210", "Engineering", "Software Engineer", "65000.00", "2026-09-24", now, now).
		AddRow(2, "Priya", "Sharma", "priya.sharma@example.com", "+919876543211", "Product", "Product Manager", "75000.00", "2026-09-25", now, now)

	mock.ExpectQuery(selectQuery).WillReturnRows(rows)

	list, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 employees, got %d", len(list))
	}
}

func TestRepositoryUpdate(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	emp := &model.Employee{
		ID:          1,
		FirstName:   "Rahul",
		LastName:    "Verma",
		Email:       "rahul.verma@example.com",
		Phone:       "+919876543210",
		Department:  "Engineering",
		Designation: "Lead Engineer",
		Salary:      decimal.NewFromInt(80000),
		JoiningDate: model.Date("2026-09-24"),
	}

	updateQuery := `UPDATE employees SET`
	mock.ExpectExec(updateQuery).
		WithArgs(
			emp.FirstName,
			emp.LastName,
			emp.Email,
			emp.Phone,
			emp.Department,
			emp.Designation,
			emp.Salary,
			"2026-09-24",
			emp.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	selectQuery := `SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at FROM employees WHERE id = ?`
	rows := sqlmock.NewRows([]string{
		"id", "first_name", "last_name", "email", "phone", "department", "designation", "salary", "joining_date", "created_at", "updated_at",
	}).AddRow(1, emp.FirstName, emp.LastName, emp.Email, emp.Phone, emp.Department, emp.Designation, "80000.00", "2026-09-24", now, now)

	mock.ExpectQuery(selectQuery).WithArgs(uint64(1)).WillReturnRows(rows)

	err := repo.Update(ctx, emp)
	if err != nil {
		t.Fatalf("unexpected error updating employee: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sql expectations: %v", err)
	}
}

func TestRepositoryDelete(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	ctx := context.Background()
	deleteQuery := `DELETE FROM employees WHERE id = ?`

	// Successful delete
	mock.ExpectExec(deleteQuery).WithArgs(uint64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Delete(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error deleting employee: %v", err)
	}

	// Delete not found
	mock.ExpectExec(deleteQuery).WithArgs(uint64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	err = repo.Delete(ctx, 999)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestRepositoryGetByEmail(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	selectQuery := `SELECT id, first_name, last_name, email, phone, department, designation, salary, joining_date, created_at, updated_at FROM employees WHERE email = ?`
	rows := sqlmock.NewRows([]string{
		"id", "first_name", "last_name", "email", "phone", "department", "designation", "salary", "joining_date", "created_at", "updated_at",
	}).AddRow(1, "Rahul", "Kumar", "rahul.kumar@example.com", "+919876543210", "Engineering", "Software Engineer", "65000.00", "2026-09-24", now, now)

	mock.ExpectQuery(selectQuery).WithArgs("rahul.kumar@example.com").WillReturnRows(rows)

	emp, err := repo.GetByEmail(ctx, "rahul.kumar@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if emp.Email != "rahul.kumar@example.com" {
		t.Errorf("expected rahul.kumar@example.com, got %s", emp.Email)
	}

	// Not found
	mock.ExpectQuery(selectQuery).WithArgs("unknown@example.com").WillReturnError(sql.ErrNoRows)
	_, err = repo.GetByEmail(ctx, "unknown@example.com")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
