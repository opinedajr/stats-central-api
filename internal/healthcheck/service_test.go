package healthcheck

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// mockDBConnection is a mock implementation of DatabaseConnection for testing
type mockDBConnection struct {
	pingError error
}

func (m *mockDBConnection) Connect(ctx context.Context) (*gorm.DB, error) {
	return nil, nil
}

func (m *mockDBConnection) Close() error {
	return nil
}

func (m *mockDBConnection) Ping() error {
	return m.pingError
}

func TestNewHealthCheckService(t *testing.T) {
	t.Run("success - creates non-nil service", func(t *testing.T) {
		mockDB := &mockDBConnection{}
		service := NewHealthCheckService(mockDB)

		if service == nil {
			t.Error("expected service to be non-nil")
		}
	})
}

func TestHealthCheckService_Check(t *testing.T) {
	t.Run("success - returns healthy status with database", func(t *testing.T) {
		mockDB := &mockDBConnection{pingError: nil}
		service := NewHealthCheckService(mockDB)
		result := service.Check()

		if len(result) != 2 {
			t.Errorf("expected 2 health check results, got %d", len(result))
		}

		// Check service health
		serviceHealth := result[0]
		if serviceHealth.ServiceName != ServiceName {
			t.Errorf("expected service name %s, got %s", ServiceName, serviceHealth.ServiceName)
		}
		if serviceHealth.Status != "healthy" {
			t.Errorf("expected status 'healthy', got %s", serviceHealth.Status)
		}
		if serviceHealth.Message != "Service is running" {
			t.Errorf("expected message 'Service is running', got %s", serviceHealth.Message)
		}

		// Check database health
		dbHealth := result[1]
		if dbHealth.ServiceName != DatabaseServiceName {
			t.Errorf("expected database service name %s, got %s", DatabaseServiceName, dbHealth.ServiceName)
		}
		if dbHealth.Status != "healthy" {
			t.Errorf("expected database status 'healthy', got %s", dbHealth.Status)
		}
		if dbHealth.Message != "Database connection successful" {
			t.Errorf("expected message 'Database connection successful', got %s", dbHealth.Message)
		}
	})

	t.Run("success - returns unhealthy status when database fails", func(t *testing.T) {
		mockErr := errors.New("connection failed")
		mockDB := &mockDBConnection{pingError: mockErr}
		service := NewHealthCheckService(mockDB)
		result := service.Check()

		assert.Equal(t, 2, len(result))

		dbHealth := result[1]
		assert.Equal(t, DatabaseServiceName, dbHealth.ServiceName)
		assert.Equal(t, "unhealthy", dbHealth.Status)
		assert.Contains(t, dbHealth.Message, "Database connection failed")
	})

	t.Run("success - handles nil database connection", func(t *testing.T) {
		service := NewHealthCheckService(nil)
		result := service.Check()

		assert.Equal(t, 2, len(result))

		dbHealth := result[1]
		assert.Equal(t, DatabaseServiceName, dbHealth.ServiceName)
		assert.Equal(t, "unhealthy", dbHealth.Status)
		assert.Contains(t, dbHealth.Message, "Database connection not initialized")
	})
}
