package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// HealthHandler handles health check requests.
type HealthHandler struct{}

// NewHealthHandler returns a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check handles GET /health
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "UP",
	})
}
