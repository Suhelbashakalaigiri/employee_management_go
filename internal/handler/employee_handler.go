package handler

import (
	"errors"
	"net/http"
	"strconv"

	apperrors "employee-management-platform/internal/errors"
	"employee-management-platform/internal/model"
	"employee-management-platform/internal/service"
	"github.com/gin-gonic/gin"
)

// EmployeeHandler handles HTTP requests for employee operations.
type EmployeeHandler struct {
	service service.EmployeeService
}

// NewEmployeeHandler creates a new EmployeeHandler.
func NewEmployeeHandler(service service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

// CreateEmployee handles POST /api/v1/employees
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req model.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.NewValidationError("Malformed JSON request payload", map[string]string{
			"body": "invalid JSON format or field types",
		}))
		return
	}

	emp, err := h.service.CreateEmployee(c.Request.Context(), &req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, emp)
}

// GetEmployeeByID handles GET /api/v1/employees/:id
func (h *EmployeeHandler) GetEmployeeByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	emp, err := h.service.GetEmployeeByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, emp)
}

// GetAllEmployees handles GET /api/v1/employees
func (h *EmployeeHandler) GetAllEmployees(c *gin.Context) {
	employees, err := h.service.GetAllEmployees(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.EmployeeListResponse{Data: employees})
}

// UpdateEmployee handles PUT /api/v1/employees/:id
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	var req model.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.NewValidationError("Malformed JSON request payload", map[string]string{
			"body": "invalid JSON format or field types",
		}))
		return
	}

	emp, err := h.service.UpdateEmployee(c.Request.Context(), id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, emp)
}

// DeleteEmployee handles DELETE /api/v1/employees/:id
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	if err := h.service.DeleteEmployee(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// parseID extracts and validates the unsigned integer :id path parameter.
func parseID(c *gin.Context) (uint64, error) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		return 0, apperrors.NewValidationError("Invalid employee ID", map[string]string{
			"id": "must be a valid positive integer",
		})
	}
	return id, nil
}

// respondError converts any error into the standard JSON error contract.
func respondError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.StatusCode, appErr.ResponsePayload())
		return
	}

	// Fallback for unexpected non-AppError
	internalErr := apperrors.NewInternalError(err)
	c.JSON(internalErr.StatusCode, internalErr.ResponsePayload())
}
