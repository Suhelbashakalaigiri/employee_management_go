package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all centralized runtime configuration for the application.
type Config struct {
	AppPort            string
	AppEnv             string
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	DBMaxOpenConns     int
	DBMaxIdleConns     int
	DBConnMaxLifetime  time.Duration
	DBConnMaxIdleTime  time.Duration
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration
}

// Load loads the configuration from environment variables, applying sensible defaults
// and validating required fields.
func Load() (*Config, error) {
	cfg := &Config{
		AppPort:            getEnv("APP_PORT", "8080"),
		AppEnv:             getEnv("APP_ENV", "development"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBName:             getEnv("DB_NAME", "employee_management"),
		DBUser:             getEnv("DB_USER", "root"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBMaxOpenConns:     getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:     getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime:  time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute,
		DBConnMaxIdleTime:  time.Duration(getEnvInt("DB_CONN_MAX_IDLE_TIME_MINUTES", 5)) * time.Minute,
		ServerReadTimeout:  10 * time.Second,
		ServerWriteTimeout: 10 * time.Second,
		ServerIdleTimeout:  60 * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	return cfg, nil
}

// Validate ensures that required configurations are present and valid.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.AppPort) == "" {
		return fmt.Errorf("APP_PORT cannot be empty")
	}
	portNum, err := strconv.Atoi(c.AppPort)
	if err != nil || portNum < 1 || portNum > 65535 {
		return fmt.Errorf("APP_PORT must be a valid port number (1-65535), got %q", c.AppPort)
	}

	if strings.TrimSpace(c.DBHost) == "" {
		return fmt.Errorf("DB_HOST cannot be empty")
	}

	if strings.TrimSpace(c.DBPort) == "" {
		return fmt.Errorf("DB_PORT cannot be empty")
	}
	dbPortNum, err := strconv.Atoi(c.DBPort)
	if err != nil || dbPortNum < 1 || dbPortNum > 65535 {
		return fmt.Errorf("DB_PORT must be a valid port number (1-65535), got %q", c.DBPort)
	}

	if strings.TrimSpace(c.DBName) == "" {
		return fmt.Errorf("DB_NAME cannot be empty")
	}

	if strings.TrimSpace(c.DBUser) == "" {
		return fmt.Errorf("DB_USER cannot be empty")
	}

	if c.DBMaxOpenConns < 1 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be at least 1, got %d", c.DBMaxOpenConns)
	}

	if c.DBMaxIdleConns < 1 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be at least 1, got %d", c.DBMaxIdleConns)
	}

	if c.DBMaxIdleConns > c.DBMaxOpenConns {
		return fmt.Errorf("DB_MAX_IDLE_CONNS (%d) cannot exceed DB_MAX_OPEN_CONNS (%d)", c.DBMaxIdleConns, c.DBMaxOpenConns)
	}

	return nil
}

// DSN returns the MySQL Data Source Name for database connection.
func (c *Config) DSN() string {
	addr := net.JoinHostPort(c.DBHost, c.DBPort)
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local",
		c.DBUser,
		c.DBPassword,
		addr,
		c.DBName,
	)
}

// ServerAddress returns the host:port string for binding the HTTP server.
func (c *Config) ServerAddress() string {
	return ":" + c.AppPort
}

func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		if intVal, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return intVal
		}
	}
	return defaultValue
}
