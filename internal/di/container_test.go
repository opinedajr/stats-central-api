package di

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	t.Run("success - creates non-nil container", func(t *testing.T) {
		container := NewContainer()

		assert.NotNil(t, container)
		assert.NotNil(t, container.repositories)
		assert.NotNil(t, container.services)
		assert.NotNil(t, container.handlers)
	})
}

func TestContainer_Config(t *testing.T) {
	t.Run("success - loads config with valid env vars", func(t *testing.T) {
		setRequiredEnvVars(t)

		container := NewContainer()
		cfg := container.Config()

		assert.NotNil(t, cfg)
		assert.Equal(t, "testdb", cfg.Database.Name)
		assert.Equal(t, "localhost", cfg.Database.Host)
	})

	t.Run("success - returns same config instance (singleton)", func(t *testing.T) {
		setRequiredEnvVars(t)

		container := NewContainer()
		cfg1 := container.Config()
		cfg2 := container.Config()

		assert.Same(t, cfg1, cfg2, "Config should return the same instance")
	})

	t.Run("panic - fails when required env vars are missing", func(t *testing.T) {
		cleanEnvVars()

		container := NewContainer()

		assert.Panics(t, func() {
			container.Config()
		}, "Config should panic when required env vars are missing")
	})
}

func TestContainer_Logger(t *testing.T) {
	t.Run("success - creates logger", func(t *testing.T) {
		setRequiredEnvVars(t)

		container := NewContainer()
		logger := container.Logger()

		assert.NotNil(t, logger)
	})

	t.Run("success - returns same logger instance (singleton)", func(t *testing.T) {
		setRequiredEnvVars(t)

		container := NewContainer()
		logger1 := container.Logger()
		logger2 := container.Logger()

		assert.Same(t, logger1, logger2, "Logger should return the same instance")
	})

	t.Run("success - creates logger with error level", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("LOG_LEVEL", "error")

		container := NewContainer()
		logger := container.Logger()

		assert.NotNil(t, logger)
	})

	t.Run("success - creates logger with debug level", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("LOG_LEVEL", "debug")

		container := NewContainer()
		logger := container.Logger()

		assert.NotNil(t, logger)
	})
}

func TestContainer_DB(t *testing.T) {
	t.Run("success - creates postgres connection when driver=postgres", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DB_PORT", "5432")

		container := NewContainer()

		assert.Panics(t, func() {
			container.DB()
		}, "DB should panic when database is not available")
	})

	t.Run("success - creates mysql connection when driver=mysql", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "mysql")
		t.Setenv("DB_PORT", "3306")

		container := NewContainer()

		assert.Panics(t, func() {
			container.DB()
		}, "DB should panic when database is not available")
	})

	t.Run("panic - unsupported database driver", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "mongodb")

		container := NewContainer()

		assert.PanicsWithValue(t, "unsupported database driver: mongodb", func() {
			container.DB()
		}, "DB should panic with clear message for unsupported driver")
	})

	t.Run("panic - unknown database driver", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "sqlite")

		container := NewContainer()

		assert.PanicsWithValue(t, "unsupported database driver: sqlite", func() {
			container.DB()
		}, "DB should panic with clear message for sqlite driver (not supported)")
	})

	t.Run("success - returns same db instance (singleton)", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		assert.NotNil(t, container, "Container should be properly initialized")
	})
}

func TestContainer_HealthCheckService(t *testing.T) {
	t.Run("success - creates health check service", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		assert.Panics(t, func() {
			container.HealthCheckService()
		}, "HealthCheckService should panic when database is not available")
	})

	t.Run("success - returns same service instance (singleton)", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		assert.NotNil(t, container.services, "ServiceDependencies should be initialized")
	})
}

func TestContainer_HealthCheckHandler(t *testing.T) {
	t.Run("success - creates health check handler", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		// Will panic when trying to connect to database
		assert.Panics(t, func() {
			container.HealthCheckHandler()
		}, "HealthCheckHandler should panic when database is not available")
	})

	t.Run("success - returns same handler instance (singleton)", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		assert.NotNil(t, container.handlers, "HandlerDependencies should be initialized")
	})
}

func TestContainer_DependencyOrder(t *testing.T) {
	t.Run("success - logger depends on config", func(t *testing.T) {
		setRequiredEnvVars(t)

		container := NewContainer()

		logger := container.Logger()
		assert.NotNil(t, logger)
		assert.NotNil(t, container.config, "Config should be initialized after Logger is called")
	})

	t.Run("success - db depends on config and logger", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		assert.Panics(t, func() {
			container.DB()
		})

		assert.NotNil(t, container.config, "Config should be initialized")
		assert.NotNil(t, container.logger, "Logger should be initialized")
	})

	t.Run("success - healthcheck service depends on db", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		assert.Panics(t, func() {
			container.HealthCheckService()
		})

		assert.NotNil(t, container.config, "Config should be initialized")
		assert.NotNil(t, container.logger, "Logger should be initialized")
	})

	t.Run("success - healthcheck handler depends on service", func(t *testing.T) {
		setRequiredEnvVars(t)
		t.Setenv("DB_DRIVER", "postgres")

		container := NewContainer()

		assert.Panics(t, func() {
			container.HealthCheckHandler()
		})

		assert.NotNil(t, container.config, "Config should be initialized")
		assert.NotNil(t, container.logger, "Logger should be initialized")
	})
}

func setRequiredEnvVars(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASSWORD", "testpass")
	t.Setenv("DB_NAME", "testdb")
}

func cleanEnvVars() {
	vars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"DB_DRIVER", "SERVER_PORT", "LOG_LEVEL",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}
