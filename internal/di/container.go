package di

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/opinedajr/stats-central-api/internal/analyse"
	"github.com/opinedajr/stats-central-api/internal/healthcheck"
	"github.com/opinedajr/stats-central-api/internal/infrastructure/database"
	"github.com/opinedajr/stats-central-api/internal/match"
	"github.com/opinedajr/stats-central-api/internal/shared/config"
	sloglogger "github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/opinedajr/stats-central-api/internal/teams"
	"github.com/opinedajr/stats-central-api/internal/tournament"
)

type Container struct {
	config       *config.Config
	logger       sloglogger.Logger
	db           *gorm.DB
	dbConn       database.DatabaseConnection
	repositories *RepositoryDependencies
	services     *ServiceDependencies
	handlers     *HandlerDependencies
}

type RepositoryDependencies struct {
	tournamentRepository tournament.Repository
	teamsRepository       teams.Repository
	matchRepository       match.Repository
	statsRepository       analyse.StatsRepository
}

type HandlerDependencies struct {
	healthcheckHandler *healthcheck.Handler
	tournamentHandler  *tournament.TournamentHandler
	teamsHandler       *teams.TeamHandler
	analyseHandler     *analyse.AnalyseHandler
}

type ServiceDependencies struct {
	healthcheckService *healthcheck.Service
	tournamentService  tournament.Service
	teamsService       teams.Service
	analyseService     analyse.Service
}

func NewContainer() *Container {
	return &Container{
		repositories: &RepositoryDependencies{},
		services:     &ServiceDependencies{},
		handlers:     &HandlerDependencies{},
	}
}

func (c *Container) Config() *config.Config {
	if c.config == nil {
		cfg, err := config.Load()
		if err != nil {
			panic("failed to load config: " + err.Error())
		}
		c.config = cfg
	}
	return c.config
}

func (c *Container) Logger() sloglogger.Logger {
	if c.logger == nil {
		cfg := c.Config()
		c.logger = sloglogger.NewLogger(cfg.Logging.Level)
	}
	return c.logger
}

func (c *Container) DB() *gorm.DB {
	if c.db == nil {
		ctx := context.Background()
		var dbConn database.DatabaseConnection

		switch c.Config().Database.Driver {
		case "mysql":
			dbConn = database.NewMySQLDatabase(c.Config().Database, c.Logger())
		case "postgres":
			dbConn = database.NewPostgresDatabase(c.Config().Database, c.Logger())
		default:
			panic(fmt.Sprintf("unsupported database driver: %s", c.Config().Database.Driver))
		}

		db, err := dbConn.Connect(ctx)
		if err != nil {
			panic("failed to connect to database: " + err.Error())
		}

		c.db = db
		c.dbConn = dbConn
	}
	return c.db
}

func (c *Container) HealthCheckService() *healthcheck.Service {
	if c.services.healthcheckService == nil {
		_ = c.DB()
		c.services.healthcheckService = healthcheck.NewHealthCheckService(c.dbConn)
	}
	return c.services.healthcheckService
}

func (c *Container) HealthCheckHandler() *healthcheck.Handler {
	if c.handlers.healthcheckHandler == nil {
		c.handlers.healthcheckHandler = healthcheck.NewHandler(c.HealthCheckService())
	}
	return c.handlers.healthcheckHandler
}

func (c *Container) TournamentRepository() tournament.Repository {
	if c.repositories.tournamentRepository == nil {
		c.repositories.tournamentRepository = tournament.NewMysqlRepository(c.DB())
	}
	return c.repositories.tournamentRepository
}

func (c *Container) TournamentService() tournament.Service {
	if c.services.tournamentService == nil {
		c.services.tournamentService = tournament.NewService(c.TournamentRepository(), c.Logger())
	}
	return c.services.tournamentService
}

func (c *Container) TournamentHandler() *tournament.TournamentHandler {
	if c.handlers.tournamentHandler == nil {
		c.handlers.tournamentHandler = tournament.NewTournamentHandler(c.TournamentService(), c.Logger())
	}
	return c.handlers.tournamentHandler
}

func (c *Container) TeamsRepository() teams.Repository {
	if c.repositories.teamsRepository == nil {
		c.repositories.teamsRepository = teams.NewMysqlRepository(c.DB())
	}
	return c.repositories.teamsRepository
}

func (c *Container) TeamsService() teams.Service {
	if c.services.teamsService == nil {
		c.services.teamsService = teams.NewService(c.TeamsRepository(), c.Logger())
	}
	return c.services.teamsService
}

func (c *Container) TeamsHandler() *teams.TeamHandler {
	if c.handlers.teamsHandler == nil {
		c.handlers.teamsHandler = teams.NewTeamHandler(c.TeamsService(), c.Logger())
	}
	return c.handlers.teamsHandler
}

func (c *Container) MatchRepository() match.Repository {
	if c.repositories.matchRepository == nil {
		c.repositories.matchRepository = match.NewMysqlRepository(c.DB())
	}
	return c.repositories.matchRepository
}

func (c *Container) StatsRepository() analyse.StatsRepository {
	if c.repositories.statsRepository == nil {
		c.repositories.statsRepository = analyse.NewMysqlStatsRepository(c.DB())
	}
	return c.repositories.statsRepository
}

func (c *Container) AnalyseService() analyse.Service {
	if c.services.analyseService == nil {
		c.services.analyseService = analyse.NewAnalyseService(c.StatsRepository(), c.MatchRepository(), c.TeamsRepository())
	}
	return c.services.analyseService
}

func (c *Container) AnalyseHandler() *analyse.AnalyseHandler {
	if c.handlers.analyseHandler == nil {
		c.handlers.analyseHandler = analyse.NewAnalyseHandler(c.AnalyseService(), c.Logger())
	}
	return c.handlers.analyseHandler
}
