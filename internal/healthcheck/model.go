package healthcheck

const ServiceName = "stats-central-api"
const ServiceMessage = "Service is running"
const DatabaseServiceName = "database"
const DatabaseMessage = "Database connection successful"
const DatabaseFailedMessage = "Database connection not initialized"
const StatusHealthy = "healthy"
const StatusUnhealthy = "unhealthy"

type Health struct {
	ServiceName string `json:"service_name"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}
