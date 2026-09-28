package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	apperrors "employee-management-platform/internal/errors"
	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics, logs the stack trace internally,
// and writes a standardized 500 error response without leaking internal details.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERED] %v\n%s", r, debug.Stack())

				errResponse := apperrors.ErrorResponse{
					Error: apperrors.ErrorDetail{
						Code:    apperrors.CodeInternalServerError,
						Message: "An unexpected error occurred",
					},
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, errResponse)
			}
		}()
		c.Next()
	}
}

// NoRouteHandler provides a standardized 404 response for unmatched routes.
func NoRouteHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, apperrors.ErrorResponse{
		Error: apperrors.ErrorDetail{
			Code:    "ROUTE_NOT_FOUND",
			Message: fmt.Sprintf("Route %s %s not found", c.Request.Method, c.Request.URL.Path),
		},
	})
}
