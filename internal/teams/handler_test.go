package teams

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
	team       *TeamOutput
	teams      []*TeamOutput
	total      int64
	findByIDErr error
	listErr    error
}

func (m *mockService) ListTeams(ctx context.Context, filter TeamFilter, page int, pageSize int) ([]*TeamOutput, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.teams, m.total, nil
}

func (m *mockService) GetTeamByID(ctx context.Context, id uint) (*TeamOutput, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.team, nil
}

func setupTestRouter(handler *TeamHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/teams", handler.List)
	r.POST("/teams", handler.MethodNotAllowed)
	r.PUT("/teams", handler.MethodNotAllowed)
	r.DELETE("/teams", handler.MethodNotAllowed)
	r.PATCH("/teams", handler.MethodNotAllowed)
	r.GET("/teams/:teamId", handler.Get)
	r.POST("/teams/:teamId", handler.MethodNotAllowed)
	r.PUT("/teams/:teamId", handler.MethodNotAllowed)
	r.DELETE("/teams/:teamId", handler.MethodNotAllowed)
	r.PATCH("/teams/:teamId", handler.MethodNotAllowed)
	return r
}

func TestHandler_List(t *testing.T) {
	logger := logger.NewLogger("info")

	tests := []struct {
		name           string
		teams          []*TeamOutput
		total          int64
		listErr        error
		queryParams    string
		expectedStatus int
		expectData     bool
	}{
		{
			name:           "success - list teams",
			teams: []*TeamOutput{
				{ID: 1, Name: "Flamengo", Country: "Brazil"},
				{ID: 2, Name: "Palmeiras", Country: "Brazil"},
			},
			total:          2,
			listErr:        nil,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name:           "success - with country filter",
			teams:          []*TeamOutput{{ID: 1, Name: "Flamengo", Country: "Brazil"}},
			total:          1,
			listErr:        nil,
			queryParams:    "?country=Brazil",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name:           "success - with name filter",
			teams:          []*TeamOutput{{ID: 1, Name: "Flamengo", Country: "Brazil"}},
			total:          1,
			listErr:        nil,
			queryParams:    "?name=Flamengo",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name:           "success - with pagination",
			teams:          []*TeamOutput{},
			total:          0,
			listErr:        nil,
			queryParams:    "?page=2&page_size=10",
			expectedStatus: http.StatusOK,
			expectData:     false,
		},
		{
			name:           "error - database error",
			teams:          nil,
			total:          0,
			listErr:        ErrDatabaseError,
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
			expectData:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockService{
				teams:   tt.teams,
				total:   tt.total,
				listErr: tt.listErr,
			}

			handler := NewTeamHandler(service, logger)
			r := setupTestRouter(handler)

			req, _ := http.NewRequest("GET", "/teams"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to parse response: %v", err)
					return
				}

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

func TestHandler_Get(t *testing.T) {
	logger := logger.NewLogger("info")

	tests := []struct {
		name           string
		team           *TeamOutput
		findByIDErr    error
		teamID         string
		expectedStatus int
	}{
		{
			name:           "success - found team",
			team:           &TeamOutput{ID: 1, Name: "Flamengo", Country: "Brazil"},
			findByIDErr:    nil,
			teamID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - invalid ID",
			team:           nil,
			findByIDErr:    nil,
			teamID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - team not found",
			team:           nil,
			findByIDErr:    ErrTeamNotFound,
			teamID:         "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "error - database error",
			team:           nil,
			findByIDErr:    ErrDatabaseError,
			teamID:         "1",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockService{
				team:       tt.team,
				findByIDErr: tt.findByIDErr,
			}

			handler := NewTeamHandler(service, logger)
			r := setupTestRouter(handler)

			req, _ := http.NewRequest("GET", "/teams/"+tt.teamID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to parse response: %v", err)
					return
				}

				if response["data"] == nil {
					t.Error("expected data in response")
				}
			}
		})
	}
}

func TestHandler_handleError(t *testing.T) {
	logger := logger.NewLogger("info")
	handler := NewTeamHandler(&mockService{}, logger)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "invalid id",
			err:            errInvalidID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_ID",
		},
		{
			name:           "team not found",
			err:            ErrTeamNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "TEAM_NOT_FOUND",
		},
		{
			name:           "validation failed",
			err:            ErrValidationFailed,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_FAILED",
		},
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

			if response["error"] == nil {
				t.Error("expected error in response")
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
	logger := logger.NewLogger("info")
	handler := NewTeamHandler(&mockService{}, logger)
	r := setupTestRouter(handler)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "POST /teams returns 405",
			method:         "POST",
			url:            "/teams",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT /teams returns 405",
			method:         "PUT",
			url:            "/teams",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE /teams returns 405",
			method:         "DELETE",
			url:            "/teams",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH /teams returns 405",
			method:         "PATCH",
			url:            "/teams",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST /teams/:id returns 405",
			method:         "POST",
			url:            "/teams/1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT /teams/:id returns 405",
			method:         "PUT",
			url:            "/teams/1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE /teams/:id returns 405",
			method:         "DELETE",
			url:            "/teams/1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH /teams/:id returns 405",
			method:         "PATCH",
			url:            "/teams/1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)
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

			if response["error"] == nil {
				t.Error("expected error in response")
				return
			}

			errorObj := response["error"].(map[string]interface{})
			if errorObj["code"] != "METHOD_NOT_ALLOWED" {
				t.Errorf("expected error code METHOD_NOT_ALLOWED, got %v", errorObj["code"])
			}
		})
	}
}
