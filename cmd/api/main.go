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

		teams := v1.Group("/teams")
		{
			teams.GET("", container.TeamsHandler().List)
			teams.POST("", container.TeamsHandler().MethodNotAllowed)
			teams.PUT("", container.TeamsHandler().MethodNotAllowed)
			teams.DELETE("", container.TeamsHandler().MethodNotAllowed)
			teams.PATCH("", container.TeamsHandler().MethodNotAllowed)

			teams.GET("/:teamId", container.TeamsHandler().Get)
			teams.POST("/:teamId", container.TeamsHandler().MethodNotAllowed)
			teams.PUT("/:teamId", container.TeamsHandler().MethodNotAllowed)
			teams.DELETE("/:teamId", container.TeamsHandler().MethodNotAllowed)
			teams.PATCH("/:teamId", container.TeamsHandler().MethodNotAllowed)
		}
	}

	log.Fatal(r.Run(":3030"))
}
