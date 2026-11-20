package analyse

import (
	"context"
	"testing"

	"github.com/opinedajr/stats-central-api/internal/match"
	"github.com/opinedajr/stats-central-api/internal/teams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStatsRepository struct {
	stats *TeamStatsEntity
	err   error
}

func (m *mockStatsRepository) GetTeamStats(ctx context.Context, teamID uint, tournamentID uint, season string) (*TeamStatsEntity, error) {
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

func (m *mockMatchRepository) GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, season string, limit int) ([]*match.MatchEntity, error) {
	if m.matchesErr != nil {
		return nil, m.matchesErr
	}
	return m.matches, nil
}

func (m *mockMatchRepository) GetHomeStats(ctx context.Context, teamID uint, tournamentID uint, season string) (match.VenueStatsEntity, error) {
	if m.homeErr != nil {
		return match.VenueStatsEntity{}, m.homeErr
	}
	return m.homeStats, nil
}

func (m *mockMatchRepository) GetAwayStats(ctx context.Context, teamID uint, tournamentID uint, season string) (match.VenueStatsEntity, error) {
	if m.awayErr != nil {
		return match.VenueStatsEntity{}, m.awayErr
	}
	return m.awayStats, nil
}

func (m *mockMatchRepository) GetOverallStats(ctx context.Context, teamID uint, tournamentID uint, season string) (match.VenueStatsEntity, error) {
	if m.overallErr != nil {
		return match.VenueStatsEntity{}, m.overallErr
	}
	return m.overallStats, nil
}

func (m *mockMatchRepository) List(ctx context.Context, filter match.MatchFilter, page int, pageSize int) ([]*match.MatchEntity, int64, error) {
	return nil, 0, nil
}

type mockTeamsRepository struct {
	team *teams.Team
	err  error
}

func (m *mockTeamsRepository) FindByID(ctx context.Context, id uint) (*teams.Team, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.team == nil {
		return &teams.Team{ID: id, Name: "Team " + string(rune(id))}, nil
	}
	return m.team, nil
}

func (m *mockTeamsRepository) List(ctx context.Context, filter teams.TeamFilter, page int, pageSize int) ([]*teams.Team, int64, error) {
	return nil, 0, nil
}

func setupTestRepos(stats *TeamStatsEntity, homeStats, awayStats, overallStats match.VenueStatsEntity, matches []*match.MatchEntity) (*mockStatsRepository, *mockMatchRepository, *mockTeamsRepository) {
	statsRepo := &mockStatsRepository{stats: stats}
	matchesRepo := &mockMatchRepository{
		homeStats:    homeStats,
		awayStats:    awayStats,
		overallStats: overallStats,
		matches:      matches,
	}
	teamsRepo := &mockTeamsRepository{}
	return statsRepo, matchesRepo, teamsRepo
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

		statsRepo, matchesRepo, teamsRepo := setupTestRepos(
			stats,
			match.VenueStatsEntity{MatchesPlayed: 1, Wins: 1, Draws: 0, Losses: 0, GoalsFor: 2, GoalsAgainst: 1},
			match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0},
			match.VenueStatsEntity{MatchesPlayed: 1, Wins: 1, Draws: 0, Losses: 0, GoalsFor: 2, GoalsAgainst: 1},
			matches,
		)

		service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		assert.Equal(t, uint(100), result.TeamID)
		assert.Equal(t, uint(1), result.TournamentID)
		assert.Equal(t, 1, result.Home.Stats.MatchesPlayed)
		assert.Equal(t, 1, result.Home.Stats.Wins)
		assert.InDelta(t, 2.5, *result.Home.Stats.AvgGoalsScored, 0.01)
		assert.Len(t, result.Overall.RecentForm, 1)
	})

	t.Run("returns ErrStatsNotFound when stats not found", func(t *testing.T) {
		statsRepo := &mockStatsRepository{err: ErrStatsNotFound}
		matchesRepo := &mockMatchRepository{}
		teamsRepo := &mockTeamsRepository{}

		service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
		_, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", 10)

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

		zeroStats := match.VenueStatsEntity{MatchesPlayed: 0}

		tests := []struct {
			name        string
			inputLastN  int
			expectError bool
		}{
			{"zero defaults to 10", 0, false},
			{"negative defaults to 10", -5, false},
			{"capped at 50", 100, false},
			{"valid within range", 25, false},
			{"boundary at 1", 1, false},
			{"boundary at 50", 50, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				statsRepo, matchesRepo, teamsRepo := setupTestRepos(stats, zeroStats, zeroStats, zeroStats, matches)

				service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
				_, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", tt.inputLastN)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
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

		zeroStats := match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0}

		statsRepo, matchesRepo, teamsRepo := setupTestRepos(stats, zeroStats, zeroStats, zeroStats, []*match.MatchEntity{})

		service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		assert.Equal(t, 0.0, result.Home.Stats.WinRate)
		assert.Equal(t, 0.0, result.Away.Stats.WinRate)
		assert.Equal(t, 0.0, result.Overall.Stats.WinRate)
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

		zeroStats := match.VenueStatsEntity{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0}

		statsRepo, matchesRepo, teamsRepo := setupTestRepos(stats, zeroStats, zeroStats, zeroStats, []*match.MatchEntity{})

		service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		assert.Empty(t, result.Overall.RecentForm)
		assert.Empty(t, result.Home.RecentForm)
		assert.Empty(t, result.Away.RecentForm)
		assert.Equal(t, 0, result.Overall.RecentFormSummary.MatchesAnalyzed)
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

		statsRepo, matchesRepo, teamsRepo := setupTestRepos(
			stats,
			match.VenueStatsEntity{MatchesPlayed: 2, Wins: 2, Draws: 0, Losses: 0, GoalsFor: 5, GoalsAgainst: 1},
			match.VenueStatsEntity{MatchesPlayed: 1, Wins: 0, Draws: 1, Losses: 0, GoalsFor: 1, GoalsAgainst: 1},
			match.VenueStatsEntity{MatchesPlayed: 3, Wins: 2, Draws: 1, Losses: 0, GoalsFor: 6, GoalsAgainst: 2},
			matches,
		)

		service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		assert.Len(t, result.Overall.RecentForm, 3)
		assert.Len(t, result.Home.RecentForm, 2)
		assert.Len(t, result.Away.RecentForm, 1)
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

		statsRepo, matchesRepo, teamsRepo := setupTestRepos(
			stats,
			match.VenueStatsEntity{MatchesPlayed: 2, Wins: 1, Draws: 1, Losses: 0, GoalsFor: 3, GoalsAgainst: 2},
			match.VenueStatsEntity{MatchesPlayed: 1, Wins: 1, Draws: 0, Losses: 0, GoalsFor: 2, GoalsAgainst: 0},
			match.VenueStatsEntity{MatchesPlayed: 3, Wins: 2, Draws: 1, Losses: 0, GoalsFor: 5, GoalsAgainst: 2},
			matches,
		)

		service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
		result, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		assert.Equal(t, 3, result.Overall.RecentFormSummary.MatchesAnalyzed)
		assert.Equal(t, 2, result.Overall.RecentFormSummary.Wins)
		assert.Equal(t, 1, result.Overall.RecentFormSummary.Draws)
		assert.Equal(t, 0, result.Overall.RecentFormSummary.Losses)
		assert.Equal(t, 3, result.Overall.RecentFormSummary.GoalsFor)
		assert.Equal(t, 4, result.Overall.RecentFormSummary.GoalsAgainst)
	})

	t.Run("returns database errors from repositories", func(t *testing.T) {
		statsRepo := &mockStatsRepository{err: ErrDatabaseError}
		matchesRepo := &mockMatchRepository{}
		teamsRepo := &mockTeamsRepository{}

		service := NewAnalyseService(statsRepo, matchesRepo, teamsRepo)
		_, err := service.TeamTournamentAnalysis(ctx, 100, 1, "2024", 10)

		assert.Error(t, err)
	})
}

func TestAnalyseService_buildRecentForm(t *testing.T) {
	timestamp := int64(1700000000)
	teamsRepo := &mockTeamsRepository{}

	tests := []struct {
		name              string
		matches           []*match.MatchEntity
		targetTeamID      uint
		limit             int
		expectedResults   []string
		expectedLength    int
	}{
		{
			name: "identifies results correctly for home team",
			matches: []*match.MatchEntity{
				{ID: 1, HomeTeamID: 100, HomeTeamGoals: 2, AwayTeamID: 200, AwayTeamGoals: 1, DateTimestamp: &timestamp},
				{ID: 2, HomeTeamID: 100, HomeTeamGoals: 1, AwayTeamID: 201, AwayTeamGoals: 1, DateTimestamp: &timestamp},
				{ID: 3, HomeTeamID: 100, HomeTeamGoals: 0, AwayTeamID: 202, AwayTeamGoals: 2, DateTimestamp: &timestamp},
			},
			targetTeamID:    100,
			limit:           10,
			expectedResults: []string{"W", "D", "L"},
			expectedLength:  3,
		},
		{
			name: "identifies results correctly for away team",
			matches: []*match.MatchEntity{
				{ID: 1, HomeTeamID: 200, HomeTeamGoals: 1, AwayTeamID: 100, AwayTeamGoals: 2, DateTimestamp: &timestamp},
				{ID: 2, HomeTeamID: 201, HomeTeamGoals: 1, AwayTeamID: 100, AwayTeamGoals: 1, DateTimestamp: &timestamp},
				{ID: 3, HomeTeamID: 202, HomeTeamGoals: 2, AwayTeamID: 100, AwayTeamGoals: 0, DateTimestamp: &timestamp},
			},
			targetTeamID:    100,
			limit:           10,
			expectedResults: []string{"W", "D", "L"},
			expectedLength:  3,
		},
		{
			name: "respects limit parameter",
			matches: func() []*match.MatchEntity {
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
				return matches
			}(),
			targetTeamID:    100,
			limit:           5,
			expectedResults: nil,
			expectedLength:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := buildRecentForm(context.Background(), tt.matches, tt.targetTeamID, tt.limit, teamsRepo)

			assert.Len(t, form, tt.expectedLength)
			if tt.expectedResults != nil {
				for i, expected := range tt.expectedResults {
					assert.Equal(t, expected, form[i].Result)
				}
			}
		})
	}
}

func TestAnalyseService_summarizeForm(t *testing.T) {
	tests := []struct {
		name            string
		form            []FormEntry
		expectedMatches int
		expectedWins    int
		expectedDraws   int
		expectedLosses  int
		expectedGoalsFor  int
		expectedGoalsAgainst int
	}{
		{
			name: "calculates summary correctly",
			form: []FormEntry{
				{Result: "W", HomeScore: 2, AwayScore: 1, Venue: "home"},
				{Result: "W", HomeScore: 3, AwayScore: 0, Venue: "home"},
				{Result: "D", HomeScore: 1, AwayScore: 1, Venue: "home"},
				{Result: "L", HomeScore: 0, AwayScore: 2, Venue: "away"},
			},
			expectedMatches:    4,
			expectedWins:      2,
			expectedDraws:     1,
			expectedLosses:    1,
			expectedGoalsFor:    6,
			expectedGoalsAgainst: 4,
		},
		{
			name: "handles empty form",
			form: []FormEntry{},
			expectedMatches:    0,
			expectedWins:      0,
			expectedDraws:     0,
			expectedLosses:    0,
			expectedGoalsFor:    0,
			expectedGoalsAgainst: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := summarizeForm(tt.form)

			assert.Equal(t, tt.expectedMatches, summary.MatchesAnalyzed)
			assert.Equal(t, tt.expectedWins, summary.Wins)
			assert.Equal(t, tt.expectedDraws, summary.Draws)
			assert.Equal(t, tt.expectedLosses, summary.Losses)
			assert.Equal(t, tt.expectedGoalsFor, summary.GoalsFor)
			assert.Equal(t, tt.expectedGoalsAgainst, summary.GoalsAgainst)
		})
	}
}

func TestAnalyseService_mapVenueStats(t *testing.T) {
	tests := []struct {
		name           string
		calcStats      match.VenueStatsEntity
		preCalcStats   *TeamStatsEntity
		venue          string
		expectedWinRate float64
	}{
		{
			name: "combines calculated and pre-calculated stats for home venue",
			calcStats: match.VenueStatsEntity{
				MatchesPlayed: 10,
				Wins:          7,
				Draws:         2,
				Losses:        1,
				GoalsFor:      22,
				GoalsAgainst:  8,
			},
			preCalcStats: func() *TeamStatsEntity {
				avgGoals := 2.5
				return &TeamStatsEntity{
					AvgGoalsScoredHome:   &avgGoals,
					AvgGoalsConcededHome: &avgGoals,
					FrequencyBTTSHome:    &avgGoals,
					FrequencyOver15Home:  &avgGoals,
				}
			}(),
			venue:          "home",
			expectedWinRate: 0.7,
		},
		{
			name: "combines calculated and pre-calculated stats for away venue",
			calcStats: match.VenueStatsEntity{
				MatchesPlayed: 10,
				Wins:          4,
				Draws:         3,
				Losses:        3,
				GoalsFor:      14,
				GoalsAgainst:  12,
			},
			preCalcStats: func() *TeamStatsEntity {
				avgGoals := 1.8
				return &TeamStatsEntity{
					AvgGoalsScoredAway:   &avgGoals,
					AvgGoalsConcededAway: &avgGoals,
					FrequencyBTTSAway:   &avgGoals,
					FrequencyOver15Away: &avgGoals,
				}
			}(),
			venue:          "away",
			expectedWinRate: 0.4,
		},
		{
			name: "combines calculated and pre-calculated stats for overall venue",
			calcStats: match.VenueStatsEntity{
				MatchesPlayed: 20,
				Wins:          11,
				Draws:         5,
				Losses:        4,
				GoalsFor:      36,
				GoalsAgainst:  20,
			},
			preCalcStats: func() *TeamStatsEntity {
				avgGoals := 2.1
				return &TeamStatsEntity{
					AvgGoalsScored:   &avgGoals,
					AvgGoalsConceded: &avgGoals,
					FrequencyBTTS:    &avgGoals,
					FrequencyOver15:  &avgGoals,
				}
			}(),
			venue:          "overall",
			expectedWinRate: 0.55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapVenueStats(tt.calcStats, tt.preCalcStats, tt.venue)

			assert.Equal(t, tt.calcStats.MatchesPlayed, result.MatchesPlayed)
			assert.Equal(t, tt.calcStats.Wins, result.Wins)
			assert.InDelta(t, tt.expectedWinRate, result.WinRate, 0.01)
			assert.NotNil(t, result.AvgGoalsScored)
		})
	}
}

func TestAnalyseService_filterHomeAndAway(t *testing.T) {
	timestamp := int64(1700000000)

	matches := []*match.MatchEntity{
		{ID: 1, HomeTeamID: 100, HomeTeamGoals: 2, AwayTeamID: 200, AwayTeamGoals: 1, DateTimestamp: &timestamp},
		{ID: 2, HomeTeamID: 300, HomeTeamGoals: 1, AwayTeamID: 100, AwayTeamGoals: 1, DateTimestamp: &timestamp},
		{ID: 3, HomeTeamID: 100, HomeTeamGoals: 3, AwayTeamID: 400, AwayTeamGoals: 0, DateTimestamp: &timestamp},
		{ID: 4, HomeTeamID: 500, HomeTeamGoals: 0, AwayTeamID: 100, AwayTeamGoals: 2, DateTimestamp: &timestamp},
	}

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
