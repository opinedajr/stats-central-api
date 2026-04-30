package healthcheck

import (
	"fmt"

	"github.com/opinedajr/stats-central-api/internal/infrastructure/database"
)

type ServiceInterface interface {
	Check() []Health
}

type Service struct {
	dbConn database.DatabaseConnection
}

func NewHealthCheckService(dbConn database.DatabaseConnection) *Service {
	return &Service{dbConn: dbConn}
}

func (s *Service) Check() []Health {
	healthChecks := []Health{
		{
			ServiceName: ServiceName,
			Status:      StatusHealthy,
			Message:     ServiceMessage,
		},
	}

	dbHealth := s.checkDatabase()
	healthChecks = append(healthChecks, dbHealth)

	return healthChecks
}

func (s *Service) checkDatabase() Health {
	if s.dbConn == nil {
		return Health{
			ServiceName: DatabaseServiceName,
			Status:      StatusUnhealthy,
			Message:     DatabaseFailedMessage,
		}
	}

	if err := s.dbConn.Ping(); err != nil {
		return Health{
			ServiceName: DatabaseServiceName,
			Status:      StatusUnhealthy,
			Message:     fmt.Sprintf("Database connection failed: %v", err),
		}
	}

	return Health{
		ServiceName: DatabaseServiceName,
		Status:      StatusHealthy,
		Message:     DatabaseMessage,
	}
}
