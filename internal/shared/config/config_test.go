package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Success(t *testing.T) {
	t.Run("success - loads config with all required env vars", func(t *testing.T) {
		t.Setenv("SERVER_PORT", "8080")
		t.Setenv("DB_DRIVER", "mysql")
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "3306")
		t.Setenv("DB_USER", "testuser")
		t.Setenv("DB_PASSWORD", "testpass")
		t.Setenv("DB_NAME", "testdb")
		t.Setenv("LOG_LEVEL", "debug")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "mysql", cfg.Database.Driver)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, "3306", cfg.Database.Port)
		assert.Equal(t, "testuser", cfg.Database.User)
		assert.Equal(t, "testpass", cfg.Database.Password)
		assert.Equal(t, "testdb", cfg.Database.Name)
		assert.Equal(t, "debug", cfg.Logging.Level)
	})

	t.Run("success - applies default values when env vars not set", func(t *testing.T) {
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_USER", "postgres")
		t.Setenv("DB_PASSWORD", "postgres")
		t.Setenv("DB_NAME", "testdb")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "3003", cfg.Server.Port)         // Default SERVER_PORT
		assert.Equal(t, "postgres", cfg.Database.Driver) // Default DB_DRIVER
		assert.Equal(t, "error", cfg.Logging.Level)      // Default LOG_LEVEL
	})

	t.Run("success - loads with postgres driver specified", func(t *testing.T) {
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_USER", "postgres")
		t.Setenv("DB_PASSWORD", "postgres")
		t.Setenv("DB_NAME", "testdb")
		t.Setenv("DB_DRIVER", "postgres")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.Equal(t, "postgres", cfg.Database.Driver)
	})

	t.Run("success - loads with mysql driver specified", func(t *testing.T) {
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "3306")
		t.Setenv("DB_USER", "mysql")
		t.Setenv("DB_PASSWORD", "mysql")
		t.Setenv("DB_NAME", "testdb")
		t.Setenv("DB_DRIVER", "mysql")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.Equal(t, "mysql", cfg.Database.Driver)
	})
}

func TestLoad_Errors(t *testing.T) {
	tests := []struct {
		name        string
		missingVars []string
		setupVars   map[string]string
		errorMsg    string
	}{
		{
			name:        "error - missing DB_HOST",
			missingVars: []string{"DB_HOST"},
			setupVars: map[string]string{
				"DB_PORT":     "5432",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "postgres",
				"DB_NAME":     "testdb",
			},
		},
		{
			name:        "error - missing DB_PORT",
			missingVars: []string{"DB_PORT"},
			setupVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "postgres",
				"DB_NAME":     "testdb",
			},
		},
		{
			name:        "error - missing DB_USER",
			missingVars: []string{"DB_USER"},
			setupVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_PASSWORD": "postgres",
				"DB_NAME":     "testdb",
			},
		},
		{
			name:        "error - missing DB_PASSWORD",
			missingVars: []string{"DB_PASSWORD"},
			setupVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DB_USER": "postgres",
				"DB_NAME": "testdb",
			},
		},
		{
			name:        "error - missing DB_NAME",
			missingVars: []string{"DB_NAME"},
			setupVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "postgres",
			},
		},
		{
			name:        "error - missing all required database vars",
			missingVars: []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"},
			setupVars:   map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for _, v := range tt.missingVars {
				os.Unsetenv(v)
			}

			for k, v := range tt.setupVars {
				t.Setenv(k, v)
			}

			cfg, err := Load()

			assert.Error(t, err)
			assert.Nil(t, cfg)
		})
	}
}

func TestConfig_Structures(t *testing.T) {
	t.Run("verify - Config struct has all required fields", func(t *testing.T) {
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_USER", "postgres")
		t.Setenv("DB_PASSWORD", "postgres")
		t.Setenv("DB_NAME", "testdb")

		cfg, err := Load()

		assert.NoError(t, err)

		assert.NotNil(t, cfg.Server)
		assert.NotNil(t, cfg.Database)
		assert.NotNil(t, cfg.Logging)
	})

	t.Run("verify - DatabaseConfig has Driver field", func(t *testing.T) {
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_USER", "postgres")
		t.Setenv("DB_PASSWORD", "postgres")
		t.Setenv("DB_NAME", "testdb")
		t.Setenv("DB_DRIVER", "mysql")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotEmpty(t, cfg.Database.Driver, "Driver field should not be empty when set")
	})
}
