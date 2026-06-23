package match

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

type mockService struct {
	matches []*MatchOutput
	total   int64
	listErr error
}

func (m *mockService) ListMatches(ctx context.Context, filter MatchFilter, page int, pageSize int) ([]*MatchOutput, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.matches, m.total, nil
}

func setupTestRouter(handler *MatchHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/matches", handler.List)
	r.POST("/matches", handler.MethodNotAllowed)
	r.PUT("/matches", handler.MethodNotAllowed)
	r.DELETE("/matches", handler.MethodNotAllowed)
	r.PATCH("/matches", handler.MethodNotAllowed)
	return r
}

func TestHandler_List(t *testing.T) {
	lg := logger.NewLogger("info")

	homeOdd := 1.85
	awayOdd := 4.20
	drawOdd := 3.40
	bttsOdd := 1.75
	under25Odd := 1.65
	homeName := "Flamengo"
	awayName := "Palmeiras"

	tests := []struct {
		name           string
		matches        []*MatchOutput
		total          int64
		listErr        error
		queryParams    string
		expectedStatus int
		expectData     bool
	}{
		{
			name: "success - list matches",
			matches: []*MatchOutput{
				{ID: 1, HomeTeamID: 100, AwayTeamID: 200, HomeTeamName: &homeName, AwayTeamName: &awayName, HomeTeamOdd: &homeOdd, AwayTeamOdd: &awayOdd, DrawOdd: &drawOdd, BTTSOdd: &bttsOdd, Under25Odd: &under25Odd},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name: "success - with tournament_id filter",
			matches: []*MatchOutput{
				{ID: 1, TournamentID: 10},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "?tournament_id=10",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name: "success - with season filter",
			matches: []*MatchOutput{
				{ID: 1, Season: "2024"},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "?season=2024",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name: "success - with status filter",
			matches: []*MatchOutput{
				{ID: 1, Status: "finished"},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "?status=finished",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name: "success - with home_team_id filter",
			matches: []*MatchOutput{
				{ID: 1, HomeTeamID: 100},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "?home_team_id=100",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name: "success - with away_team_id filter",
			matches: []*MatchOutput{
				{ID: 1, AwayTeamID: 200},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "?away_team_id=200",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name: "success - with round filter",
			matches: []*MatchOutput{
				{ID: 1, Round: 15},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "?round=15",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name: "success - with pagination",
			matches:        []*MatchOutput{},
			total:          0,
			listErr:        nil,
			queryParams:    "?page=2&page_size=10",
			expectedStatus: http.StatusOK,
			expectData:     false,
		},
		{
			name: "success - combined filters",
			matches: []*MatchOutput{
				{ID: 1},
			},
			total:          1,
			listErr:        nil,
			queryParams:    "?tournament_id=10&season=2024&status=finished&round=15&home_team_id=100&away_team_id=200",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name:           "error - database error",
			matches:        nil,
			total:          0,
			listErr:        ErrDatabaseError,
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
			expectData:     false,
		},
		{
			name:           "error - unexpected error",
			matches:        nil,
			total:          0,
			listErr:        errors.New("unknown"),
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
			expectData:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				matches: tt.matches,
				total:   tt.total,
				listErr: tt.listErr,
			}

			handler := NewMatchHandler(svc, lg)
			r := setupTestRouter(handler)

			req, _ := http.NewRequest("GET", "/matches"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("failed to parse response: %v", err)
				return
			}

			if tt.expectedStatus == http.StatusOK {
				if tt.expectData && response["data"] == nil {
					t.Error("expected data in response")
				}

				if response["meta"] == nil {
					t.Error("expected meta in response")
				}
			}
		})
	}
}

func TestHandler_handleError(t *testing.T) {
	lg := logger.NewLogger("info")
	handler := NewMatchHandler(&mockService{}, lg)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "database error",
			err:            ErrDatabaseError,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "DATABASE_ERROR",
		},
		{
			name:           "unknown error",
			err:            errors.New("unknown"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/error", func(c *gin.Context) {
				handler.handleError(c, tt.err)
			})

			req, _ := http.NewRequest("GET", "/error", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("failed to parse response: %v", err)
				return
			}

			errorObj := response["error"].(map[string]interface{})
			if errorObj["code"] != tt.expectedCode {
				t.Errorf("expected error code %s, got %v", tt.expectedCode, errorObj["code"])
			}
		})
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	lg := logger.NewLogger("info")
	handler := NewMatchHandler(&mockService{}, lg)
	r := setupTestRouter(handler)

	tests := []struct {
		name   string
		method string
	}{
		{name: "POST returns 405", method: "POST"},
		{name: "PUT returns 405", method: "PUT"},
		{name: "DELETE returns 405", method: "DELETE"},
		{name: "PATCH returns 405", method: "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, "/matches", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("failed to parse response: %v", err)
				return
			}

			errorObj := response["error"].(map[string]interface{})
			if errorObj["code"] != "METHOD_NOT_ALLOWED" {
				t.Errorf("expected error code METHOD_NOT_ALLOWED, got %v", errorObj["code"])
			}
		})
	}
}
