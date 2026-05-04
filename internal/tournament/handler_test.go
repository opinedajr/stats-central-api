package tournament

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/shared/api"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTournamentServiceForHandler struct {
	mock.Mock
}

func (m *MockTournamentServiceForHandler) CreateTournament(ctx context.Context, input CreateTournamentInput) (*TournamentOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TournamentOutput), args.Error(1)
}

func (m *MockTournamentServiceForHandler) ListTournaments(ctx context.Context, filter TournamentFilter, page int, pageSize int) ([]*TournamentOutput, int64, error) {
	args := m.Called(ctx, filter, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*TournamentOutput), args.Get(1).(int64), args.Error(2)
}

func (m *MockTournamentServiceForHandler) GetTournamentByID(ctx context.Context, id uint) (*TournamentOutput, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TournamentOutput), args.Error(1)
}

func (m *MockTournamentServiceForHandler) UpdateTournament(ctx context.Context, id uint, input UpdateTournamentInput) (*TournamentOutput, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TournamentOutput), args.Error(1)
}

func (m *MockTournamentServiceForHandler) UpdateTournamentStatus(ctx context.Context, id uint, active bool) (*TournamentOutput, error) {
	args := m.Called(ctx, id, active)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TournamentOutput), args.Error(1)
}

func setupTestRouter(handler *TournamentHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestCreateTournamentHandler(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	division := 1
	season := "2024-2025"
	active := true

	tests := []struct {
		name           string
		requestBody    any
		expectedStatus int
		setupMock      func(*MockTournamentServiceForHandler)
		assertResponse func(*testing.T, []byte)
		assertMock     func(*testing.T, *MockTournamentServiceForHandler)
	}{
		{
			name: "success",
			requestBody: CreateTournamentInput{
				Name:     "Premier League",
				Country:  "England",
				Division: &division,
				Season:   &season,
				Active:   &active,
			},
			expectedStatus: http.StatusCreated,
			setupMock: func(m *MockTournamentServiceForHandler) {
				expectedOutput := &TournamentOutput{
					ID:        1,
					Name:      "Premier League",
					Country:   "England",
					Division:  &division,
					Season:    &season,
					Active:    active,
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}
				m.On("CreateTournament", mock.Anything, mock.MatchedBy(func(input CreateTournamentInput) bool {
					return input.Name == "Premier League" &&
						input.Country == "England" &&
						input.Division != nil &&
						*input.Division == 1 &&
						input.Season != nil &&
						*input.Season == "2024-2025" &&
						input.Active != nil &&
						*input.Active == true
				})).Return(expectedOutput, nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				assert.Equal(t, uint(1), response.Data.ID)
				assert.Equal(t, "Premier League", response.Data.Name)
				assert.Equal(t, "England", response.Data.Country)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
		{
			name:           "validation error - invalid request body",
			requestBody:    []byte("{invalid json"),
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(m *MockTournamentServiceForHandler) {},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertNotCalled(t, "CreateTournament")
			},
		},
		{
			name: "service error",
			requestBody: CreateTournamentInput{
				Name:    "Test Tournament",
				Country: "England",
			},
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *MockTournamentServiceForHandler) {
				m.On("CreateTournament", mock.Anything, mock.AnythingOfType("tournament.CreateTournamentInput")).Return(nil, ErrDatabaseError).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "DATABASE_ERROR", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTournamentServiceForHandler)
			logger := logger.NewLogger("error")
			handler := NewTournamentHandler(mockService, logger)

			tt.setupMock(mockService)

			var bodyBytes []byte
			var err error

			switch b := tt.requestBody.(type) {
			case []byte:
				bodyBytes = b
			case CreateTournamentInput:
				bodyBytes, err = json.Marshal(b)
				require.NoError(t, err)
			}

			router := setupTestRouter(handler)
			req, _ := http.NewRequest("POST", "/api/v1/tournaments", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.POST("/api/v1/tournaments", handler.Create)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.assertResponse(t, w.Body.Bytes())
			tt.assertMock(t, mockService)
		})
	}
}

func TestListTournamentsHandler(t *testing.T) {
	division1 := 1
	division2 := 2
	season := "2024-2025"
	active := true

	tests := []struct {
		name           string
		queryString    string
		expectedStatus int
		setupMock      func(*MockTournamentServiceForHandler)
		assertResponse func(*testing.T, []byte)
		assertMock     func(*testing.T, *MockTournamentServiceForHandler)
	}{
		{
			name:           "success",
			queryString:    "?page=1&page_size=20",
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockTournamentServiceForHandler) {
				expectedOutputs := []*TournamentOutput{
					{
						ID:       2,
						Name:     "Championship",
						Country:  "England",
						Division: &division2,
						Season:   &season,
						Active:   true,
					},
					{
						ID:       1,
						Name:     "Premier League",
						Country:  "England",
						Division: &division1,
						Season:   &season,
						Active:   true,
					},
				}
				m.On("ListTournaments", mock.Anything, TournamentFilter{}, 1, 20).Return(expectedOutputs, int64(2), nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[[]*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				assert.Len(t, response.Data, 2)
				assert.Equal(t, "Championship", response.Data[0].Name)
				assert.Equal(t, "Premier League", response.Data[1].Name)
				assert.NotNil(t, response.Meta)
				assert.Equal(t, int64(2), response.Meta.Total)
				assert.Equal(t, 1, response.Meta.Page)
				assert.Equal(t, 20, response.Meta.PageSize)
				assert.Equal(t, 1, response.Meta.TotalPages)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
		{
			name:           "success with filters",
			queryString:    "?active=true&country=England&division=1&season=2024-2025",
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockTournamentServiceForHandler) {
				country := "England"
				expectedOutputs := []*TournamentOutput{
					{
						ID:       1,
						Name:     "Premier League",
						Country:  "England",
						Division: &division1,
						Season:   &season,
						Active:   true,
					},
				}
				expectedFilter := TournamentFilter{
					Active:   &active,
					Country:  &country,
					Division: &division1,
					Season:   &season,
				}
				m.On("ListTournaments", mock.Anything, expectedFilter, 1, 20).Return(expectedOutputs, int64(1), nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[[]*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.Len(t, response.Data, 1)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
		{
			name:           "empty list",
			queryString:    "",
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockTournamentServiceForHandler) {
				m.On("ListTournaments", mock.Anything, TournamentFilter{}, 1, 20).Return([]*TournamentOutput{}, int64(0), nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[[]*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.Len(t, response.Data, 0)
				assert.Equal(t, int64(0), response.Meta.Total)
				assert.Equal(t, 1, response.Meta.TotalPages)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTournamentServiceForHandler)
			logger := logger.NewLogger("error")
			handler := NewTournamentHandler(mockService, logger)

			tt.setupMock(mockService)

			router := setupTestRouter(handler)
			req, _ := http.NewRequest("GET", "/api/v1/tournaments"+tt.queryString, nil)
			w := httptest.NewRecorder()

			router.GET("/api/v1/tournaments", handler.List)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.assertResponse(t, w.Body.Bytes())
			tt.assertMock(t, mockService)
		})
	}
}

func TestGetTournamentHandler(t *testing.T) {
	division := 1
	season := "2024-2025"

	tests := []struct {
		name           string
		tournamentID   string
		expectedStatus int
		setupMock      func(*MockTournamentServiceForHandler)
		assertResponse func(*testing.T, []byte)
		assertMock     func(*testing.T, *MockTournamentServiceForHandler)
	}{
		{
			name:           "success",
			tournamentID:   "123",
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockTournamentServiceForHandler) {
				expectedOutput := &TournamentOutput{
					ID:       123,
					Name:     "Premier League",
					Country:  "England",
					Division: &division,
					Season:   &season,
					Active:   true,
				}
				m.On("GetTournamentByID", mock.Anything, uint(123)).Return(expectedOutput, nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				assert.Equal(t, uint(123), response.Data.ID)
				assert.Equal(t, "Premier League", response.Data.Name)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
		{
			name:           "invalid id",
			tournamentID:   "invalid",
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(m *MockTournamentServiceForHandler) {},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "INVALID_ID", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertNotCalled(t, "GetTournamentByID")
			},
		},
		{
			name:           "tournament not found",
			tournamentID:   "99999",
			expectedStatus: http.StatusNotFound,
			setupMock: func(m *MockTournamentServiceForHandler) {
				m.On("GetTournamentByID", mock.Anything, uint(99999)).Return(nil, ErrTournamentNotFound).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "TOURNAMENT_NOT_FOUND", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTournamentServiceForHandler)
			logger := logger.NewLogger("error")
			handler := NewTournamentHandler(mockService, logger)

			tt.setupMock(mockService)

			router := setupTestRouter(handler)
			req, _ := http.NewRequest("GET", "/api/v1/tournaments/"+tt.tournamentID, nil)
			w := httptest.NewRecorder()

			router.GET("/api/v1/tournaments/:tournamentId", handler.Get)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.assertResponse(t, w.Body.Bytes())
			tt.assertMock(t, mockService)
		})
	}
}

func TestUpdateTournamentHandler(t *testing.T) {
	division := 1
	season := "2024-2025"

	tests := []struct {
		name           string
		tournamentID   string
		requestBody    UpdateTournamentInput
		expectedStatus int
		setupMock      func(*MockTournamentServiceForHandler)
		assertResponse func(*testing.T, []byte)
		assertMock     func(*testing.T, *MockTournamentServiceForHandler)
	}{
		{
			name:         "success",
			tournamentID: "123",
			requestBody: UpdateTournamentInput{
				Name:     "Updated Premier League",
				Country:  "England",
				Division: &division,
				Season:   &season,
			},
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockTournamentServiceForHandler) {
				expectedOutput := &TournamentOutput{
					ID:       123,
					Name:     "Updated Premier League",
					Country:  "England",
					Division: &division,
					Season:   &season,
					Active:   true,
				}
				m.On("UpdateTournament", mock.Anything, uint(123), mock.MatchedBy(func(input UpdateTournamentInput) bool {
					return input.Name == "Updated Premier League" &&
						input.Country == "England" &&
						input.Division != nil &&
						*input.Division == 1 &&
						input.Season != nil &&
						*input.Season == "2024-2025"
				})).Return(expectedOutput, nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				assert.Equal(t, "Updated Premier League", response.Data.Name)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
		{
			name:         "invalid id",
			tournamentID: "invalid",
			requestBody: UpdateTournamentInput{
				Name:    "Updated",
				Country: "England",
			},
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(m *MockTournamentServiceForHandler) {},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "INVALID_ID", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertNotCalled(t, "UpdateTournament")
			},
		},
		{
			name:         "tournament not found",
			tournamentID: "99999",
			requestBody: UpdateTournamentInput{
				Name:    "Updated",
				Country: "England",
			},
			expectedStatus: http.StatusNotFound,
			setupMock: func(m *MockTournamentServiceForHandler) {
				m.On("UpdateTournament", mock.Anything, uint(99999), mock.AnythingOfType("tournament.UpdateTournamentInput")).Return(nil, ErrTournamentNotFound).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "TOURNAMENT_NOT_FOUND", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTournamentServiceForHandler)
			logger := logger.NewLogger("error")
			handler := NewTournamentHandler(mockService, logger)

			tt.setupMock(mockService)

			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			router := setupTestRouter(handler)
			req, _ := http.NewRequest("PUT", "/api/v1/tournaments/"+tt.tournamentID, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.PUT("/api/v1/tournaments/:tournamentId", handler.Update)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.assertResponse(t, w.Body.Bytes())
			tt.assertMock(t, mockService)
		})
	}
}

func TestUpdateTournamentStatusHandler(t *testing.T) {
	tests := []struct {
		name           string
		tournamentID   string
		active         bool
		expectedStatus int
		setupMock      func(*MockTournamentServiceForHandler)
		assertResponse func(*testing.T, []byte)
		assertMock     func(*testing.T, *MockTournamentServiceForHandler)
	}{
		{
			name:           "success activate",
			tournamentID:   "123",
			active:         true,
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockTournamentServiceForHandler) {
				expectedOutput := &TournamentOutput{
					ID:      123,
					Name:    "Premier League",
					Country: "England",
					Active:  true,
				}
				m.On("UpdateTournamentStatus", mock.Anything, uint(123), true).Return(expectedOutput, nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				assert.True(t, response.Data.Active)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
		{
			name:           "success deactivate",
			tournamentID:   "123",
			active:         false,
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockTournamentServiceForHandler) {
				expectedOutput := &TournamentOutput{
					ID:      123,
					Name:    "Premier League",
					Country: "England",
					Active:  false,
				}
				m.On("UpdateTournamentStatus", mock.Anything, uint(123), false).Return(expectedOutput, nil).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[*TournamentOutput]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Data)
				assert.False(t, response.Data.Active)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
		{
			name:           "invalid id",
			tournamentID:   "invalid",
			active:         true,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(m *MockTournamentServiceForHandler) {},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "INVALID_ID", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertNotCalled(t, "UpdateTournamentStatus")
			},
		},
		{
			name:           "tournament not found",
			tournamentID:   "99999",
			active:         true,
			expectedStatus: http.StatusNotFound,
			setupMock: func(m *MockTournamentServiceForHandler) {
				m.On("UpdateTournamentStatus", mock.Anything, uint(99999), true).Return(nil, ErrTournamentNotFound).Once()
			},
			assertResponse: func(t *testing.T, body []byte) {
				var response api.Response[struct{}]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Error)
				assert.Equal(t, "TOURNAMENT_NOT_FOUND", response.Error.Code)
			},
			assertMock: func(t *testing.T, m *MockTournamentServiceForHandler) {
				m.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTournamentServiceForHandler)
			logger := logger.NewLogger("error")
			handler := NewTournamentHandler(mockService, logger)

			tt.setupMock(mockService)

			requestBody := UpdateTournamentStatusInput{
				Active: tt.active,
			}

			bodyBytes, err := json.Marshal(requestBody)
			require.NoError(t, err)

			router := setupTestRouter(handler)
			req, _ := http.NewRequest("PATCH", "/api/v1/tournaments/"+tt.tournamentID+"/status", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.PATCH("/api/v1/tournaments/:tournamentId/status", handler.UpdateStatus)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.assertResponse(t, w.Body.Bytes())
			tt.assertMock(t, mockService)
		})
	}
}
