package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"employee-management-platform/internal/config"
	"employee-management-platform/internal/database"
	"employee-management-platform/internal/handler"
	"employee-management-platform/internal/repository"
	"employee-management-platform/internal/router"
	"employee-management-platform/internal/service"
)

func main() {
	log.Println("[STARTUP] Starting Employee Management Platform V1...")

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[STARTUP ERROR] Failed to load configuration: %v", err)
	}

	// 2. Initialize database & connection pool
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[STARTUP ERROR] Database connection failed: %v", err)
	}

	// 3. Initialize layers (Repository -> Service -> Handler -> Router)
	empRepo := repository.NewEmployeeRepository(db)
	empService := service.NewEmployeeService(empRepo)
	empHandler := handler.NewEmployeeHandler(empService)
	healthHandler := handler.NewHealthHandler()

	r := router.SetupRouter(cfg, empHandler, healthHandler)

	// 4. Configure HTTP Server
	srv := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      r,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}

	// 5. Start HTTP server in a separate goroutine
	go func() {
		log.Printf("[SERVER] Listening and serving HTTP on %s (Environment: %s)", cfg.ServerAddress(), cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[SERVER ERROR] HTTP server failed to start: %v", err)
		}
	}()

	// 6. Graceful shutdown on termination signals (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("[SHUTDOWN] Received signal %v, initiating graceful shutdown...", sig)

	// Context for graceful HTTP shutdown (allow active requests up to 10 seconds to finish)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop accepting new requests and close HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[SHUTDOWN] HTTP server shutdown error: %v", err)
	} else {
		log.Println("[SHUTDOWN] HTTP server closed successfully.")
	}

	// Close database resources
	if err := database.Close(db); err != nil {
		log.Printf("[SHUTDOWN] Database close error: %v", err)
	} else {
		log.Println("[SHUTDOWN] Database connections closed successfully.")
	}

	log.Println("[SHUTDOWN] Application stopped cleanly.")
}
