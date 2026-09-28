package config_test

import (
	"os"
	"testing"
	"time"

	"employee-management-platform/internal/config"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Clear relevant env vars
	os.Unsetenv("APP_PORT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error loading defaults, got: %v", err)
	}

	if cfg.AppPort != "8080" {
		t.Errorf("expected AppPort 8080, got %s", cfg.AppPort)
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("expected DBHost localhost, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "3306" {
		t.Errorf("expected DBPort 3306, got %s", cfg.DBPort)
	}
	if cfg.DBName != "employee_management" {
		t.Errorf("expected DBName employee_management, got %s", cfg.DBName)
	}
	if cfg.DBUser != "root" {
		t.Errorf("expected DBUser root, got %s", cfg.DBUser)
	}
	if cfg.ServerAddress() != ":8080" {
		t.Errorf("expected ServerAddress :8080, got %s", cfg.ServerAddress())
	}
	if cfg.DBMaxOpenConns != 25 {
		t.Errorf("expected DBMaxOpenConns 25, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBConnMaxLifetime != 5*time.Minute {
		t.Errorf("expected DBConnMaxLifetime 5m, got %v", cfg.DBConnMaxLifetime)
	}
}

func TestConfigCustomValues(t *testing.T) {
	t.Setenv("APP_PORT", "9000")
	t.Setenv("DB_HOST", "mysql-host")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_NAME", "custom_db")
	t.Setenv("DB_USER", "appuser")
	t.Setenv("DB_PASSWORD", "secret123")
	t.Setenv("DB_MAX_OPEN_CONNS", "50")
	t.Setenv("DB_MAX_IDLE_CONNS", "20")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.AppPort != "9000" {
		t.Errorf("expected AppPort 9000, got %s", cfg.AppPort)
	}
	if cfg.DBHost != "mysql-host" {
		t.Errorf("expected DBHost mysql-host, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "3307" {
		t.Errorf("expected DBPort 3307, got %s", cfg.DBPort)
	}
	if cfg.DBName != "custom_db" {
		t.Errorf("expected DBName custom_db, got %s", cfg.DBName)
	}
	if cfg.DBUser != "appuser" {
		t.Errorf("expected DBUser appuser, got %s", cfg.DBUser)
	}
	if cfg.DBPassword != "secret123" {
		t.Errorf("expected DBPassword secret123, got %s", cfg.DBPassword)
	}
	if cfg.DBMaxOpenConns != 50 {
		t.Errorf("expected DBMaxOpenConns 50, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 20 {
		t.Errorf("expected DBMaxIdleConns 20, got %d", cfg.DBMaxIdleConns)
	}

	expectedDSN := "appuser:secret123@tcp(mysql-host:3307)/custom_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local"
	if cfg.DSN() != expectedDSN {
		t.Errorf("expected DSN %s, got %s", expectedDSN, cfg.DSN())
	}
}

func TestConfigValidationFailures(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		envVal  string
		wantErr bool
	}{
		{
			name:    "invalid app port string",
			envKey:  "APP_PORT",
			envVal:  "invalid-port",
			wantErr: true,
		},
		{
			name:    "port out of range",
			envKey:  "APP_PORT",
			envVal:  "70000",
			wantErr: true,
		},
		{
			name:    "invalid db port",
			envKey:  "DB_PORT",
			envVal:  "0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envKey, tt.envVal)
			_, err := config.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("config.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
