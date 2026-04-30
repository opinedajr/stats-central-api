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

	log.Fatal(r.Run(":3030"))
}
