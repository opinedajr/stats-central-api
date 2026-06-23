package match

import "context"

type MatchFilter struct {
	TournamentID *uint
	Season       *string
	Round        *int
	Status       *string
	HomeTeamID   *uint
	AwayTeamID   *uint
}

type Repository interface {
	List(ctx context.Context, filter MatchFilter, page int, pageSize int) ([]*MatchEntity, int64, error)
	GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, season string, limit int) ([]*MatchEntity, error)
	GetHomeStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error)
	GetAwayStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error)
	GetOverallStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error)
}
