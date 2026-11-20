package healthcheck

import (
	"fmt"

	"github.com/opinedajr/stats-central-api/internal/infrastructure/database"
)

type Service interface {
	Check() []Health
}

type service struct {
	dbConn database.DatabaseConnection
}

var _ Service = (*service)(nil)

func NewHealthCheckService(dbConn database.DatabaseConnection) Service {
	return &service{dbConn: dbConn}
}

func (s *service) Check() []Health {
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

func (s *service) checkDatabase() Health {
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
