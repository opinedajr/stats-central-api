package database

import (
	"context"
	"testing"

	"github.com/opinedajr/stats-central-api/internal/shared/config"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewPostgresDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      config.DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "error - invalid host",
			config: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "invalid-host-that-does-not-exist-123456789",
				Port:     "5432",
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
				Driver:   "postgres",
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
				Driver:   "postgres",
				Host:     "",
				Port:     "5432",
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
			pgDB := NewPostgresDatabase(tt.config, log)
			db, err := pgDB.Connect(ctx)

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

func TestPostgresDatabase_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		config config.DatabaseConfig
	}{
		{
			name: "error - connection with valid config structure fails",
			config: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     "5432",
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
			pgDB := NewPostgresDatabase(tt.config, log)
			db, err := pgDB.Connect(ctx)

			assert.Error(t, err)
			assert.Nil(t, db)

			if err != nil {
				assert.Contains(t, err.Error(), "failed to connect to database")
			}
		})
	}
}

func TestPostgresDatabase_Migrate(t *testing.T) {
	t.Run("success - migrate returns nil (no-op)", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		pgDB := NewPostgresDatabase(cfg, log)

		type TestModel struct {
			ID   uint
			Name string
		}

		err := pgDB.Migrate(&TestModel{})

		assert.NoError(t, err)
	})
}

func TestPostgresDatabase_Close(t *testing.T) {
	t.Run("success - close without connection returns nil", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		pgDB := NewPostgresDatabase(cfg, log)
		err := pgDB.Close()

		assert.NoError(t, err)
	})
}

func TestPostgresDatabase_Ping(t *testing.T) {
	t.Run("error - ping without connection returns error", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		pgDB := NewPostgresDatabase(cfg, log)
		err := pgDB.Ping()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not connected")
	})
}
