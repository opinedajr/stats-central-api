package analyse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/shared/api"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAnalyseService struct {
	output AnalyseOutput
	err    error
}

func (m *mockAnalyseService) TeamTournamentAnalysis(ctx context.Context, teamID uint, tournamentID uint, season string, lastN int) (AnalyseOutput, error) {
	if m.err != nil {
		return AnalyseOutput{}, m.err
	}
	return m.output, nil
}

func setupTestRouter(handler *AnalyseHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAnalyseHandler_TeamTournamentAnalysis(t *testing.T) {
	t.Run("returns 200 with complete analysis on success", func(t *testing.T) {
		avgGoals := 2.5
		output := AnalyseOutput{
			TeamID:       100,
			TournamentID: 1,
			Home: VenueContext{
				Stats: VenueStats{
					MatchesPlayed:     10,
					Wins:              7,
					Draws:             2,
					Losses:            1,
					GoalsFor:          22,
					GoalsAgainst:      8,
					WinRate:           0.7,
					AvgGoalsScored:    &avgGoals,
					AvgGoalsConceded:  &avgGoals,
					FrequencyBTTS:     &avgGoals,
					FrequencyOver15:   &avgGoals,
				},
				RecentForm: []FormEntry{
					{MatchID: 1, Result: "W", HomeID: 100, HomeName: "Team Home", AwayID: 200, AwayName: "Team A", HomeScore: 2, AwayScore: 1, Venue: "home", Date: time.Now()},
				},
				RecentFormSummary: FormSummary{
					MatchesAnalyzed: 1,
					Wins:            1,
					Draws:           0,
					Losses:          0,
					GoalsFor:        2,
					GoalsAgainst:    1,
				},
			},
			Away:         VenueContext{},
			Overall: VenueContext{
				RecentForm: []FormEntry{
					{MatchID: 1, Result: "W", HomeID: 100, HomeName: "Team Home", AwayID: 200, AwayName: "Team A", HomeScore: 2, AwayScore: 1, Venue: "home", Date: time.Now()},
				},
			},
			CalculatedAt: time.Now(),
		}

		mockService := &mockAnalyseService{output: output}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?tournament_id=1&season=2024&last_n=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[AnalyseOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.Data)
		assert.Equal(t, uint(100), response.Data.TeamID)
		assert.Equal(t, uint(1), response.Data.TournamentID)
		assert.Equal(t, 10, response.Data.Home.Stats.MatchesPlayed)
		assert.Equal(t, 7, response.Data.Home.Stats.Wins)
		assert.Len(t, response.Data.Overall.RecentForm, 1)
	})

	t.Run("returns 404 when stats not found", func(t *testing.T) {
		mockService := &mockAnalyseService{err: ErrStatsNotFound}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?tournament_id=1&season=2024", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[struct{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.Error)
		assert.Equal(t, "NOT_FOUND", response.Error.Code)
	})

	t.Run("returns 404 when team not found", func(t *testing.T) {
		mockService := &mockAnalyseService{err: ErrTeamNotFound}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?tournament_id=1&season=2024", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 404 when tournament not found", func(t *testing.T) {
		mockService := &mockAnalyseService{err: ErrTournamentNotFound}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?tournament_id=1&season=2024", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for invalid team ID", func(t *testing.T) {
		mockService := &mockAnalyseService{}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/invalid/analysis?tournament_id=1&season=2024", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[struct{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.Error)
		assert.Equal(t, "INVALID_ID", response.Error.Code)
	})

	t.Run("returns 400 for invalid tournament ID", func(t *testing.T) {
		mockService := &mockAnalyseService{}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?season=2024", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[struct{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.Error)
		assert.Equal(t, "INVALID_ID", response.Error.Code)
	})

	t.Run("parses last_n query parameter correctly", func(t *testing.T) {
		avgGoals := 2.0
		output := AnalyseOutput{
			TeamID:       100,
			TournamentID: 1,
			Home: VenueContext{
				Stats: VenueStats{
					MatchesPlayed:     10,
					Wins:              7,
					Draws:             2,
					Losses:            1,
					GoalsFor:          22,
					GoalsAgainst:      8,
					WinRate:           0.7,
					AvgGoalsScored:    &avgGoals,
				},
				RecentForm: []FormEntry{
					{MatchID: 1, Result: "W", HomeID: 100, HomeName: "Team Home", AwayID: 200, AwayName: "Team A", HomeScore: 2, AwayScore: 1, Venue: "home", Date: time.Now()},
				},
				RecentFormSummary: FormSummary{
					MatchesAnalyzed: 1,
					Wins:            1,
					Draws:           0,
					Losses:          0,
					GoalsFor:        2,
					GoalsAgainst:    1,
				},
			},
			Away:         VenueContext{},
			Overall: VenueContext{
				RecentForm: []FormEntry{
					{MatchID: 1, Result: "W", HomeID: 100, HomeName: "Team Home", AwayID: 200, AwayName: "Team A", HomeScore: 2, AwayScore: 1, Venue: "home", Date: time.Now()},
				},
			},
			CalculatedAt: time.Now(),
		}

		mockService := &mockAnalyseService{output: output}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?tournament_id=1&season=2024&last_n=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("uses default last_n=10 when not provided", func(t *testing.T) {
		avgGoals := 2.0
		output := AnalyseOutput{
			TeamID:       100,
			TournamentID: 1,
			Home: VenueContext{
				Stats: VenueStats{MatchesPlayed: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0, WinRate: 0, AvgGoalsScored: &avgGoals},
			},
			Away:         VenueContext{},
			Overall: VenueContext{
				RecentForm: []FormEntry{
					{MatchID: 1, Result: "W", HomeID: 100, HomeName: "Team Home", AwayID: 200, AwayName: "Team A", HomeScore: 2, AwayScore: 1, Venue: "home", Date: time.Now()},
				},
			},
			CalculatedAt: time.Now(),
		}

		mockService := &mockAnalyseService{output: output}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?tournament_id=1&season=2024", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 405 for non-GET requests", func(t *testing.T) {
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(&mockAnalyseService{}, log)
		router := setupTestRouter(handler)
		router.POST("/analyse/team/:teamId/analysis", handler.MethodNotAllowed)

		req, _ := http.NewRequest("POST", "/analyse/team/100/analysis", bytes.NewBuffer([]byte("{}")))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

		var response api.Response[struct{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.Error)
		assert.Equal(t, "METHOD_NOT_ALLOWED", response.Error.Code)
	})

	t.Run("returns 500 for unexpected errors", func(t *testing.T) {
		mockService := &mockAnalyseService{err: errors.New("unexpected error")}
		log := logger.NewLogger("info")
		handler := NewAnalyseHandler(mockService, log)
		router := setupTestRouter(handler)
		router.GET("/analyse/team/:teamId/analysis", handler.TeamTournamentAnalysis)

		req, _ := http.NewRequest("GET", "/analyse/team/100/analysis?tournament_id=1&season=2024", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[struct{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.Error)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
