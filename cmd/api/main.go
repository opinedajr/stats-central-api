package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/di"
)

func main() {
	container := di.NewContainer()
	r := gin.Default()

	r.GET("/health", container.HealthCheckHandler().Handle)

	v1 := r.Group("/v1")
	{
		tournaments := v1.Group("/tournaments")
		{
			tournaments.POST("", container.TournamentHandler().Create)
			tournaments.GET("", container.TournamentHandler().List)
			tournaments.GET("/:tournamentId", container.TournamentHandler().Get)
			tournaments.PUT("/:tournamentId", container.TournamentHandler().Update)
			tournaments.PATCH("/:tournamentId/status", container.TournamentHandler().UpdateStatus)
		}
	}

	log.Fatal(r.Run(":3030"))
}
