package match

import "context"

type Repository interface {
	GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, season string, limit int) ([]*MatchEntity, error)
	GetHomeStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error)
	GetAwayStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error)
	GetOverallStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error)
}
