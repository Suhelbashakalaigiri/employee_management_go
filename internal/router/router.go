package router

import (
	"employee-management-platform/internal/config"
	"employee-management-platform/internal/handler"
	"employee-management-platform/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin engine with all routes and middlewares configured.
func SetupRouter(
	cfg *config.Config,
	empHandler *handler.EmployeeHandler,
	healthHandler *handler.HealthHandler,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Apply core middlewares
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Custom 404 handler matching standard API error schema
	r.NoRoute(middleware.NoRouteHandler)

	// Health endpoint outside /api/v1 as per API design
	r.GET("/health", healthHandler.Check)

	// API v1 business endpoints
	v1 := r.Group("/api/v1")
	{
		employees := v1.Group("/employees")
		{
			employees.POST("", empHandler.CreateEmployee)
			employees.GET("", empHandler.GetAllEmployees)
			employees.GET("/:id", empHandler.GetEmployeeByID)
			employees.PUT("/:id", empHandler.UpdateEmployee)
			employees.DELETE("/:id", empHandler.DeleteEmployee)
		}
	}

	return r
}
