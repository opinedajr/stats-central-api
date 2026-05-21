package analyse

import "context"

type StatsRepository interface {
	GetTeamStats(ctx context.Context, teamID uint, tournamentID uint) (*TeamStatsEntity, error)
}
