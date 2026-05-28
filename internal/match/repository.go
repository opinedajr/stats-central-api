package match

import "context"

type Repository interface {
	GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, limit int) ([]*MatchEntity, error)
	GetHomeStats(ctx context.Context, teamID uint, tournamentID uint) (VenueStatsEntity, error)
	GetAwayStats(ctx context.Context, teamID uint, tournamentID uint) (VenueStatsEntity, error)
	GetOverallStats(ctx context.Context, teamID uint, tournamentID uint) (VenueStatsEntity, error)
}
