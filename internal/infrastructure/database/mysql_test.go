package database

import (
	"context"
	"testing"

	"github.com/opinedajr/stats-central-api/internal/shared/config"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewMySQLDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      config.DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "error - invalid host",
			config: config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "invalid-host-that-does-not-exist-123456789",
				Port:     "3306",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError: true,
			errorMsg:    "failed to connect to database",
		},
		{
			name: "error - invalid port",
			config: config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     "invalid-port",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError: true,
			errorMsg:    "failed to connect to database",
		},
		{
			name: "error - empty host",
			config: config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "",
				Port:     "3306",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError: true,
			errorMsg:    "failed to connect to database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewLogger("error")
			ctx := context.Background()
			mysqlDB := NewMySQLDatabase(tt.config, log)
			db, err := mysqlDB.Connect(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}

func TestMySQLDatabase_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		config config.DatabaseConfig
	}{
		{
			name: "error - connection with valid config structure fails",
			config: config.DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     "3306",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewLogger("error")
			ctx := context.Background()
			mysqlDB := NewMySQLDatabase(tt.config, log)
			db, err := mysqlDB.Connect(ctx)

			assert.Error(t, err)
			assert.Nil(t, db)

			if err != nil {
				assert.Contains(t, err.Error(), "failed to connect to database")
			}
		})
	}
}

func TestMySQLDatabase_Migrate(t *testing.T) {
	t.Run("error - migrate without connection returns error", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     "3306",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		mysqlDB := NewMySQLDatabase(cfg, log)

		type TestModel struct {
			ID   uint
			Name string
		}

		err := mysqlDB.Migrate(&TestModel{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not connected")
	})
}

func TestMySQLDatabase_Close(t *testing.T) {
	t.Run("success - close without connection returns nil", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     "3306",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		mysqlDB := NewMySQLDatabase(cfg, log)
		err := mysqlDB.Close()

		assert.NoError(t, err)
	})
}

func TestMySQLDatabase_Ping(t *testing.T) {
	t.Run("error - ping without connection returns error", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     "3306",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		mysqlDB := NewMySQLDatabase(cfg, log)
		err := mysqlDB.Ping()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not connected")
	})
}
