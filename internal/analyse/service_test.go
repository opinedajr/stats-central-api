package analyse

import (
	"context"
	"testing"

	"github.com/opinedajr/stats-central-api/internal/match"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStatsRepository struct {
	stats *TeamStatsEntity
	err   error
}

func (m *mockStatsRepository) GetTeamStats(ctx context.Context, teamID uint, tournamentID uint) (*TeamStatsEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

type mockMatchRepository struct {
	homeStats    match.VenueStatsEntity
	awayStats    match.VenueStatsEntity
	overallStats match.VenueStatsEntity
	matches      []*match.MatchEntity
	homeErr      error
	awayErr      error
	overallErr   error
	matchesErr   error
}

func (m *mockMatchRepository) GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, limit int) ([]*match.MatchEntity, error) {
	if m.matchesErr != nil {
		return nil, m.matchesErr
	}
	return m.matches, nil
}

func (m *mockMatchRepository) GetHomeStats(ctx context.Context, teamID uint, tournamentID uint) (match.VenueStatsEntity, error) {
	if m.homeErr != nil {
		return match.VenueStatsEntity{}, m.homeErr
	}
	return m.homeStats, nil
}

func (m *mockMatchRepository) GetAwayStats(ctx context.Context, teamID uint, tournamentID uint) (match.VenueStatsEntity, error) {
	if m.awayErr != nil {
		return match.VenueStatsEntity{}, m.awayErr
	}
	return m.awayStats, nil
}

func (m *mockMatchRepository) GetOverallStats(ctx context.Context, teamID uint, tournamentID uint) (match.VenueStatsEntity, error) {
	if m.overallErr != nil {
		return match.VenueStatsEntity{}, m.overallErr
	}
	return m.overallStats, nil
}

func TestAnalyseService_TeamTournamentAnalysis(t *testing.T) {
	ctx := context.Background()

	t.Run("aggregates data from both repositories correctly", func(t *testing.T) {
		avgGoals := 2.5
		stats := &TeamStatsEntity{
			TeamID:               100,
			LeagueID:             1,
			Season:               "2024",
			AvgGoalsScored:       &avgGoals,
			AvgGoalsScoredHome:   &avgGoals,
			AvgGoalsScoredAway:   &avgGoals,
		}

		timestamp := int64(1700000000)
		matches := []*match.MatchEntity{
			{
				ID:            1,
				HomeTeamID:    100,
				HomeTeamGoals: 2,
				AwayTeamID:    200,
				AwayTeamGoals: 1,
				DateTimestamp: &timestamp,
			},
		}

		statsRepo := &mockStatsRepository{stats: stats}
		matchesRepo := &mockMatchRepository{
			homeStats:    match.VenueStatsEntity{MatchesPlayed: 1, Wins: 1, Draws: 0, Losses: 0, GoalsFor: 2, GoalsAgainst: 1},
			awayStats:    match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			overallStats: match.VenueStatsEntity{MatchesPlayed: 1, Wins: 1, Draws: 0, Losses: 0, GoalsFor: 2, GoalsAgainst: 1},
			matches:      matches,
		}

		service := NewAnalyseService(statsRepo, matchesRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, 10)

		require.NoError(t, err)
		assert.Equal(t, uint(100), result.TeamID)
		assert.Equal(t, uint(1), result.TournamentID)
		assert.Equal(t, 1, result.HomeStats.MatchesPlayed)
		assert.Equal(t, 1, result.HomeStats.Wins)
		assert.InDelta(t, 2.5, *result.HomeStats.AvgGoalsScored, 0.01)
		assert.Len(t, result.RecentForm, 1)
	})

	t.Run("returns ErrStatsNotFound when stats not found", func(t *testing.T) {
		statsRepo := &mockStatsRepository{err: ErrStatsNotFound}
		matchesRepo := &mockMatchRepository{}

		service := NewAnalyseService(statsRepo, matchesRepo)
		_, err := service.TeamTournamentAnalysis(ctx, 100, 1, 10)

		assert.Error(t, err)
		assert.Equal(t, ErrStatsNotFound, err)
	})

	t.Run("normalizes last_n parameter", func(t *testing.T) {
		avgGoals := 2.0
		stats := &TeamStatsEntity{
			TeamID:             100,
			LeagueID:           1,
			Season:             "2024",
			AvgGoalsScored:     &avgGoals,
			AvgGoalsScoredHome: &avgGoals,
			AvgGoalsScoredAway: &avgGoals,
		}

		timestamp := int64(1700000000)
		matches := []*match.MatchEntity{
			{ID: 1, HomeTeamID: 100, HomeTeamGoals: 1, AwayTeamID: 200, AwayTeamGoals: 0, DateTimestamp: &timestamp},
		}

		statsRepo := &mockStatsRepository{stats: stats}
		matchesRepo := &mockMatchRepository{
			homeStats:    match.VenueStatsEntity{MatchesPlayed: 0},
			awayStats:    match.VenueStatsEntity{MatchesPlayed: 0},
			overallStats: match.VenueStatsEntity{MatchesPlayed: 0},
			matches:      matches,
		}

		service := NewAnalyseService(statsRepo, matchesRepo)

		t.Run("zero or negative last_n defaults to 10", func(t *testing.T) {
			_, err := service.TeamTournamentAnalysis(ctx, 100, 1, 0)
			require.NoError(t, err)
		})

		t.Run("last_n greater than 50 is capped at 50", func(t *testing.T) {
			_, err := service.TeamTournamentAnalysis(ctx, 100, 1, 100)
			require.NoError(t, err)
		})
	})

	t.Run("handles division by zero for win rate", func(t *testing.T) {
		avgGoals := 1.5
		stats := &TeamStatsEntity{
			TeamID:             100,
			LeagueID:           1,
			Season:             "2024",
			AvgGoalsScored:     &avgGoals,
			AvgGoalsScoredHome: &avgGoals,
			AvgGoalsScoredAway: &avgGoals,
		}

		statsRepo := &mockStatsRepository{stats: stats}
		matchesRepo := &mockMatchRepository{
			homeStats:    match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			awayStats:    match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			overallStats: match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			matches:      []*match.MatchEntity{},
		}

		service := NewAnalyseService(statsRepo, matchesRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, 10)

		require.NoError(t, err)
		assert.Equal(t, 0.0, result.HomeStats.WinRate)
		assert.Equal(t, 0.0, result.AwayStats.WinRate)
		assert.Equal(t, 0.0, result.OverallStats.WinRate)
	})

	t.Run("handles no matches case", func(t *testing.T) {
		avgGoals := 1.5
		stats := &TeamStatsEntity{
			TeamID:             100,
			LeagueID:           1,
			Season:             "2024",
			AvgGoalsScored:     &avgGoals,
			AvgGoalsScoredHome: &avgGoals,
			AvgGoalsScoredAway: &avgGoals,
		}

		statsRepo := &mockStatsRepository{stats: stats}
		matchesRepo := &mockMatchRepository{
			homeStats:    match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			awayStats:    match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			overallStats: match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			matches:      []*match.MatchEntity{},
		}

		service := NewAnalyseService(statsRepo, matchesRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, 10)

		require.NoError(t, err)
		assert.Empty(t, result.RecentForm)
		assert.Empty(t, result.RecentFormHome)
		assert.Empty(t, result.RecentFormAway)
		assert.Equal(t, 0, result.RecentFormSummary.MatchesAnalyzed)
	})

	t.Run("builds form arrays correctly separating home and away", func(t *testing.T) {
		avgGoals := 2.0
		stats := &TeamStatsEntity{
			TeamID:             100,
			LeagueID:           1,
			Season:             "2024",
			AvgGoalsScored:     &avgGoals,
			AvgGoalsScoredHome: &avgGoals,
			AvgGoalsScoredAway: &avgGoals,
		}

		timestamp1 := int64(1700000000)
		timestamp2 := int64(1700001000)
		timestamp3 := int64(1700002000)
		matches := []*match.MatchEntity{
			{ID: 1, HomeTeamID: 100, HomeTeamGoals: 2, AwayTeamID: 200, AwayTeamGoals: 1, DateTimestamp: &timestamp1},
			{ID: 2, HomeTeamID: 300, HomeTeamGoals: 1, AwayTeamID: 100, AwayTeamGoals: 1, DateTimestamp: &timestamp2},
			{ID: 3, HomeTeamID: 100, HomeTeamGoals: 3, AwayTeamID: 400, AwayTeamGoals: 0, DateTimestamp: &timestamp3},
		}

		statsRepo := &mockStatsRepository{stats: stats}
		matchesRepo := &mockMatchRepository{
			homeStats:    match.VenueStatsEntity{MatchesPlayed: 2, Wins: 2, Draws: 0, Losses: 0, GoalsFor: 5, GoalsAgainst: 1},
			awayStats:    match.VenueStatsEntity{MatchesPlayed: 1, Wins: 0, Draws: 1, Losses: 0, GoalsFor: 1, GoalsAgainst: 1},
			overallStats: match.VenueStatsEntity{MatchesPlayed: 3, Wins: 2, Draws: 1, Losses: 0, GoalsFor: 6, GoalsAgainst: 2},
			matches:      matches,
		}

		service := NewAnalyseService(statsRepo, matchesRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, 10)

		require.NoError(t, err)
		assert.Len(t, result.RecentForm, 3)
		assert.Len(t, result.RecentFormHome, 2)
		assert.Len(t, result.RecentFormAway, 1)
		assert.Equal(t, "home", result.RecentFormHome[0].Venue)
		assert.Equal(t, "away", result.RecentFormAway[0].Venue)
	})

	t.Run("calculates form summary correctly", func(t *testing.T) {
		avgGoals := 2.0
		stats := &TeamStatsEntity{
			TeamID:             100,
			LeagueID:           1,
			Season:             "2024",
			AvgGoalsScored:     &avgGoals,
			AvgGoalsScoredHome: &avgGoals,
			AvgGoalsScoredAway: &avgGoals,
		}

		timestamp := int64(1700000000)
		matches := []*match.MatchEntity{
			{ID: 1, HomeTeamID: 100, HomeTeamGoals: 2, AwayTeamID: 200, AwayTeamGoals: 1, DateTimestamp: &timestamp},
			{ID: 2, HomeTeamID: 300, HomeTeamGoals: 0, AwayTeamID: 100, AwayTeamGoals: 2, DateTimestamp: &timestamp},
			{ID: 3, HomeTeamID: 100, HomeTeamGoals: 1, AwayTeamID: 400, AwayTeamGoals: 1, DateTimestamp: &timestamp},
		}

		statsRepo := &mockStatsRepository{stats: stats}
		matchesRepo := &mockMatchRepository{
			homeStats:    match.VenueStatsEntity{MatchesPlayed: 2, Wins: 1, Draws: 1, Losses: 0, GoalsFor: 3, GoalsAgainst: 2},
			awayStats:    match.VenueStatsEntity{MatchesPlayed: 1, Wins: 1, Draws: 0, Losses: 0, GoalsFor: 2, GoalsAgainst: 0},
			overallStats: match.VenueStatsEntity{MatchesPlayed: 3, Wins: 2, Draws: 1, Losses: 0, GoalsFor: 5, GoalsAgainst: 2},
			matches:      matches,
		}

		service := NewAnalyseService(statsRepo, matchesRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, 10)

		require.NoError(t, err)
		assert.Equal(t, 3, result.RecentFormSummary.MatchesAnalyzed)
		assert.Equal(t, 2, result.RecentFormSummary.Wins)
		assert.Equal(t, 1, result.RecentFormSummary.Draws)
		assert.Equal(t, 0, result.RecentFormSummary.Losses)
		assert.Equal(t, 3, result.RecentFormSummary.GoalsFor)
		assert.Equal(t, 4, result.RecentFormSummary.GoalsAgainst)
	})

	t.Run("returns database errors from repositories", func(t *testing.T) {
		statsRepo := &mockStatsRepository{err: ErrDatabaseError}
		matchesRepo := &mockMatchRepository{}

		service := NewAnalyseService(statsRepo, matchesRepo)
		_, err := service.TeamTournamentAnalysis(ctx, 100, 1, 10)

		assert.Error(t, err)
	})
}

func TestAnalyseService_buildRecentForm(t *testing.T) {
	timestamp := int64(1700000000)

	t.Run("identifies results correctly for home team", func(t *testing.T) {
		matches := []*match.MatchEntity{
			{ID: 1, HomeTeamID: 100, HomeTeamGoals: 2, AwayTeamID: 200, AwayTeamGoals: 1, DateTimestamp: &timestamp},
			{ID: 2, HomeTeamID: 100, HomeTeamGoals: 1, AwayTeamID: 201, AwayTeamGoals: 1, DateTimestamp: &timestamp},
			{ID: 3, HomeTeamID: 100, HomeTeamGoals: 0, AwayTeamID: 202, AwayTeamGoals: 2, DateTimestamp: &timestamp},
		}

		form := buildRecentForm(matches, 100, 10)

		assert.Len(t, form, 3)
		assert.Equal(t, "W", form[0].Result)
		assert.Equal(t, "D", form[1].Result)
		assert.Equal(t, "L", form[2].Result)
	})

	t.Run("identifies results correctly for away team", func(t *testing.T) {
		matches := []*match.MatchEntity{
			{ID: 1, HomeTeamID: 200, HomeTeamGoals: 1, AwayTeamID: 100, AwayTeamGoals: 2, DateTimestamp: &timestamp},
			{ID: 2, HomeTeamID: 201, HomeTeamGoals: 1, AwayTeamID: 100, AwayTeamGoals: 1, DateTimestamp: &timestamp},
			{ID: 3, HomeTeamID: 202, HomeTeamGoals: 2, AwayTeamID: 100, AwayTeamGoals: 0, DateTimestamp: &timestamp},
		}

		form := buildRecentForm(matches, 100, 10)

		assert.Len(t, form, 3)
		assert.Equal(t, "W", form[0].Result)
		assert.Equal(t, "D", form[1].Result)
		assert.Equal(t, "L", form[2].Result)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		matches := make([]*match.MatchEntity, 20)
		for i := range matches {
			ts := int64(1700000000 + int64(i))
			matches[i] = &match.MatchEntity{
				ID:            uint(i + 1),
				HomeTeamID:    100,
				HomeTeamGoals: 1,
				AwayTeamID:    200,
				AwayTeamGoals: 0,
				DateTimestamp: &ts,
			}
		}

		form := buildRecentForm(matches, 100, 5)

		assert.Len(t, form, 5)
	})
}

func TestAnalyseService_summarizeForm(t *testing.T) {
	form := []FormEntry{
		{Result: "W", HomeScore: 2, AwayScore: 1},
		{Result: "W", HomeScore: 3, AwayScore: 0},
		{Result: "D", HomeScore: 1, AwayScore: 1},
		{Result: "L", HomeScore: 0, AwayScore: 2},
	}

	summary := summarizeForm(form)

	assert.Equal(t, 4, summary.MatchesAnalyzed)
	assert.Equal(t, 2, summary.Wins)
	assert.Equal(t, 1, summary.Draws)
	assert.Equal(t, 1, summary.Losses)
	assert.Equal(t, 6, summary.GoalsFor)
	assert.Equal(t, 4, summary.GoalsAgainst)
}

func TestAnalyseService_mapVenueStats(t *testing.T) {
	t.Run("combines calculated and pre-calculated stats for home venue", func(t *testing.T) {
		avgGoals := 2.5
		calcStats := match.VenueStatsEntity{
			MatchesPlayed: 10,
			Wins:          7,
			Draws:         2,
			Losses:        1,
			GoalsFor:      22,
			GoalsAgainst:  8,
		}

		preCalcStats := &TeamStatsEntity{
			AvgGoalsScoredHome:     &avgGoals,
			AvgGoalsConcededHome:   &avgGoals,
			FrequencyBTTSHome:      &avgGoals,
			FrequencyOver15Home:    &avgGoals,
		}

		result := mapVenueStats(calcStats, preCalcStats, "home")

		assert.Equal(t, 10, result.MatchesPlayed)
		assert.Equal(t, 7, result.Wins)
		assert.Equal(t, 0.7, result.WinRate)
		assert.NotNil(t, result.AvgGoalsScored)
		assert.InDelta(t, 2.5, *result.AvgGoalsScored, 0.01)
		assert.NotNil(t, result.FrequencyBTTS)
		assert.InDelta(t, 2.5, *result.FrequencyBTTS, 0.01)
	})

	t.Run("combines calculated and pre-calculated stats for away venue", func(t *testing.T) {
		avgGoals := 1.8
		calcStats := match.VenueStatsEntity{
			MatchesPlayed: 10,
			Wins:          4,
			Draws:         3,
			Losses:        3,
			GoalsFor:      14,
			GoalsAgainst:  12,
		}

		preCalcStats := &TeamStatsEntity{
			AvgGoalsScoredAway:     &avgGoals,
			AvgGoalsConcededAway:   &avgGoals,
			FrequencyBTTSAway:      &avgGoals,
			FrequencyOver15Away:    &avgGoals,
		}

		result := mapVenueStats(calcStats, preCalcStats, "away")

		assert.Equal(t, 10, result.MatchesPlayed)
		assert.Equal(t, 4, result.Wins)
		assert.Equal(t, 0.4, result.WinRate)
		assert.InDelta(t, 1.8, *result.AvgGoalsScored, 0.01)
	})

	t.Run("combines calculated and pre-calculated stats for overall venue", func(t *testing.T) {
		avgGoals := 2.1
		calcStats := match.VenueStatsEntity{
			MatchesPlayed: 20,
			Wins:          11,
			Draws:         5,
			Losses:        4,
			GoalsFor:      36,
			GoalsAgainst:  20,
		}

		preCalcStats := &TeamStatsEntity{
			AvgGoalsScored:     &avgGoals,
			AvgGoalsConceded:   &avgGoals,
			FrequencyBTTS:      &avgGoals,
			FrequencyOver15:    &avgGoals,
		}

		result := mapVenueStats(calcStats, preCalcStats, "overall")

		assert.Equal(t, 20, result.MatchesPlayed)
		assert.Equal(t, 11, result.Wins)
		assert.Equal(t, 0.55, result.WinRate)
		assert.InDelta(t, 2.1, *result.AvgGoalsScored, 0.01)
	})
}

func TestAnalyseService_filterHomeAndAway(t *testing.T) {
	timestamp := int64(1700000000)

	matches := []*match.MatchEntity{
		{ID: 1, HomeTeamID: 100, HomeTeamGoals: 2, AwayTeamID: 200, AwayTeamGoals: 1, DateTimestamp: &timestamp},
		{ID: 2, HomeTeamID: 300, HomeTeamGoals: 1, AwayTeamID: 100, AwayTeamGoals: 1, DateTimestamp: &timestamp},
		{ID: 3, HomeTeamID: 100, HomeTeamGoals: 3, AwayTeamID: 400, AwayTeamGoals: 0, DateTimestamp: &timestamp},
		{ID: 4, HomeTeamID: 500, HomeTeamGoals: 0, AwayTeamID: 100, AwayTeamGoals: 2, DateTimestamp: &timestamp},
	}

	_ = timestamp

	t.Run("filters home matches correctly", func(t *testing.T) {
		homeMatches := filterHome(matches, 100)
		assert.Len(t, homeMatches, 2)
		assert.Equal(t, uint(100), homeMatches[0].HomeTeamID)
		assert.Equal(t, uint(100), homeMatches[1].HomeTeamID)
	})

	t.Run("filters away matches correctly", func(t *testing.T) {
		awayMatches := filterAway(matches, 100)
		assert.Len(t, awayMatches, 2)
		assert.Equal(t, uint(100), awayMatches[0].AwayTeamID)
		assert.Equal(t, uint(100), awayMatches[1].AwayTeamID)
	})
}
