package tournament

import (
	"context"
	"testing"
	"time"

	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, tournament *Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockRepository) Update(ctx context.Context, tournament *Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id uint, active bool) error {
	args := m.Called(ctx, id, active)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id uint) (*Tournament, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Tournament), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, filter TournamentFilter, page int, pageSize int) ([]*Tournament, int64, error) {
	args := m.Called(ctx, filter, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Tournament), args.Get(1).(int64), args.Error(2)
}

func TestCreateTournament(t *testing.T) {
	division := 1
	season := "2024-2025"
	active := true

	tests := []struct {
		name         string
		input        CreateTournamentInput
		setupMock    func(*MockRepository)
		wantErr      error
		assertOutput func(*testing.T, *TournamentOutput)
	}{
		{
			name: "success",
			input: CreateTournamentInput{
				Name:     "Premier League",
				Country:  "England",
				Division: &division,
				Season:   &season,
				Active:   &active,
			},
			setupMock: func(m *MockRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*tournament.Tournament")).Return(nil).Once()
			},
			wantErr: nil,
			assertOutput: func(tt *testing.T, output *TournamentOutput) {
				assert.Equal(tt, "Premier League", output.Name)
				assert.Equal(tt, "England", output.Country)
				assert.True(tt, output.Active)
			},
		},
		{
			name: "success with nil optional fields",
			input: CreateTournamentInput{
				Name:    "La Liga",
				Country: "Spain",
			},
			setupMock: func(m *MockRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*tournament.Tournament")).Return(nil).Once()
			},
			wantErr: nil,
			assertOutput: func(tt *testing.T, output *TournamentOutput) {
				assert.Equal(tt, "La Liga", output.Name)
				assert.Equal(tt, "Spain", output.Country)
				assert.Nil(tt, output.Division)
				assert.Nil(tt, output.Season)
				assert.True(tt, output.Active)
			},
		},
		{
			name: "repository error",
			input: CreateTournamentInput{
				Name:    "Test Tournament",
				Country: "England",
			},
			setupMock: func(m *MockRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*tournament.Tournament")).Return(ErrDatabaseError).Once()
			},
			wantErr: ErrDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			logger := logger.NewLogger("error")
			service := NewService(mockRepo, logger)

			ctx := context.Background()
			tt.setupMock(mockRepo)

			output, err := service.CreateTournament(ctx, tt.input)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, output)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				if tt.assertOutput != nil {
					tt.assertOutput(t, output)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListTournaments(t *testing.T) {
	division1 := 1
	division2 := 2
	season1 := "2024-2025"
	season2 := "2024-2025"

	tournaments := []*Tournament{
		{
			ID:       2,
			Name:     "Championship",
			Country:  "England",
			Division: &division2,
			Season:   &season2,
			Active:   true,
		},
		{
			ID:       1,
			Name:     "Premier League",
			Country:  "England",
			Division: &division1,
			Season:   &season1,
			Active:   true,
		},
	}

	tests := []struct {
		name          string
		filter        TournamentFilter
		page          int
		pageSize      int
		setupMock     func(*MockRepository)
		wantErr       bool
		wantTotal     int64
		wantLen       int
		assertOutputs func(*testing.T, []*TournamentOutput)
	}{
		{
			name:     "success",
			filter:   TournamentFilter{},
			page:     1,
			pageSize: 20,
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, TournamentFilter{}, 1, 20).Return(tournaments, int64(2), nil).Once()
			},
			wantErr:   false,
			wantTotal: 2,
			wantLen:   2,
			assertOutputs: func(tt *testing.T, outputs []*TournamentOutput) {
				assert.Equal(tt, "Championship", outputs[0].Name)
				assert.Equal(tt, "Premier League", outputs[1].Name)
			},
		},
		{
			name:     "empty list",
			filter:   TournamentFilter{},
			page:     1,
			pageSize: 20,
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, TournamentFilter{}, 1, 20).Return([]*Tournament{}, int64(0), nil).Once()
			},
			wantErr:   false,
			wantTotal: 0,
			wantLen:   0,
		},
		{
			name:     "default page and page_size applied when not provided",
			filter:   TournamentFilter{},
			page:     0,
			pageSize: 0,
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, TournamentFilter{}, 1, 20).Return([]*Tournament{}, int64(0), nil).Once()
			},
			wantErr:   false,
			wantTotal: 0,
			wantLen:   0,
		},
		{
			name:     "page_size capped at 100",
			filter:   TournamentFilter{},
			page:     1,
			pageSize: 200,
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, TournamentFilter{}, 1, 100).Return([]*Tournament{}, int64(0), nil).Once()
			},
			wantErr:   false,
			wantTotal: 0,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			logger := logger.NewLogger("error")
			service := NewService(mockRepo, logger)

			ctx := context.Background()
			tt.setupMock(mockRepo)

			outputs, total, err := service.ListTournaments(ctx, tt.filter, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, outputs, tt.wantLen)
			if tt.assertOutputs != nil && len(outputs) > 0 {
				tt.assertOutputs(t, outputs)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetTournamentByID(t *testing.T) {
	division := 1
	season := "2024-2025"

	tournament := &Tournament{
		ID:       123,
		Name:     "Premier League",
		Country:  "England",
		Division: &division,
		Season:   &season,
		Active:   true,
	}

	tests := []struct {
		name         string
		tournamentID uint
		setupMock    func(*MockRepository)
		wantErr      error
		assertOutput func(*testing.T, *TournamentOutput)
	}{
		{
			name:         "success maps to TournamentOutput",
			tournamentID: 123,
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(123)).Return(tournament, nil).Once()
			},
			wantErr: nil,
			assertOutput: func(tt *testing.T, output *TournamentOutput) {
				assert.Equal(tt, uint(123), output.ID)
				assert.Equal(tt, "Premier League", output.Name)
				assert.Equal(tt, "England", output.Country)
				assert.Equal(tt, 1, *output.Division)
				assert.Equal(tt, "2024-2025", *output.Season)
				assert.True(tt, output.Active)
			},
		},
		{
			name:         "repo returns ErrTournamentNotFound propagates",
			tournamentID: 99999,
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(99999)).Return(nil, ErrTournamentNotFound).Once()
			},
			wantErr: ErrTournamentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			logger := logger.NewLogger("error")
			service := NewService(mockRepo, logger)

			ctx := context.Background()
			tt.setupMock(mockRepo)

			output, err := service.GetTournamentByID(ctx, tt.tournamentID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, output)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				if tt.assertOutput != nil {
					tt.assertOutput(t, output)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateTournament(t *testing.T) {
	existingDivision := 1
	existingSeason := "2024-2025"
	updatedDivision := 2
	updatedSeason := "2023-2024"

	existingTournament := &Tournament{
		ID:       123,
		Name:     "Original Name",
		Country:  "England",
		Division: &existingDivision,
		Season:   &existingSeason,
		Active:   true,
	}

	updatedTournament := &Tournament{
		ID:       123,
		Name:     "Updated Premier League",
		Country:  "England",
		Division: &updatedDivision,
		Season:   &updatedSeason,
		Active:   true,
	}

	tests := []struct {
		name         string
		tournamentID uint
		input        UpdateTournamentInput
		setupMock    func(*MockRepository)
		wantErr      error
		assertOutput func(*testing.T, *TournamentOutput)
	}{
		{
			name:         "success maps input to domain model calls repo returns updated TournamentOutput",
			tournamentID: 123,
			input: UpdateTournamentInput{
				Name:     "Updated Premier League",
				Country:  "England",
				Division: &updatedDivision,
				Season:   &updatedSeason,
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(123)).Return(existingTournament, nil).Once()
				m.On("Update", mock.Anything, mock.AnythingOfType("*tournament.Tournament")).Return(nil).Once()
				m.On("FindByID", mock.Anything, uint(123)).Return(updatedTournament, nil).Once()
			},
			wantErr: nil,
			assertOutput: func(tt *testing.T, output *TournamentOutput) {
				assert.Equal(tt, "Updated Premier League", output.Name)
				assert.Equal(tt, 2, *output.Division)
				assert.Equal(tt, "2023-2024", *output.Season)
			},
		},
		{
			name:         "repo returns ErrTournamentNotFound propagated",
			tournamentID: 99999,
			input: UpdateTournamentInput{
				Name:    "Updated Tournament",
				Country: "England",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(99999)).Return(nil, ErrTournamentNotFound).Once()
			},
			wantErr: ErrTournamentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			logger := logger.NewLogger("error")
			service := NewService(mockRepo, logger)

			ctx := context.Background()
			tt.setupMock(mockRepo)

			output, err := service.UpdateTournament(ctx, tt.tournamentID, tt.input)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, output)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				if tt.assertOutput != nil {
					tt.assertOutput(t, output)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateTournamentStatus(t *testing.T) {
	activatedTournament := &Tournament{
		ID:      123,
		Name:    "Premier League",
		Country: "England",
		Active:  true,
	}

	deactivatedTournament := &Tournament{
		ID:      123,
		Name:    "Premier League",
		Country: "England",
		Active:  false,
	}

	tests := []struct {
		name         string
		tournamentID uint
		active       bool
		setupMock    func(*MockRepository)
		wantErr      error
		assertOutput func(*testing.T, *TournamentOutput)
	}{
		{
			name:         "success activate calls repo.UpdateStatus with correct args returns updated TournamentOutput",
			tournamentID: 123,
			active:       true,
			setupMock: func(m *MockRepository) {
				m.On("UpdateStatus", mock.Anything, uint(123), true).Return(nil).Once()
				m.On("FindByID", mock.Anything, uint(123)).Return(activatedTournament, nil).Once()
			},
			wantErr: nil,
			assertOutput: func(tt *testing.T, output *TournamentOutput) {
				assert.True(tt, output.Active)
			},
		},
		{
			name:         "success deactivate calls repo.UpdateStatus with correct args returns updated TournamentOutput",
			tournamentID: 123,
			active:       false,
			setupMock: func(m *MockRepository) {
				m.On("UpdateStatus", mock.Anything, uint(123), false).Return(nil).Once()
				m.On("FindByID", mock.Anything, uint(123)).Return(deactivatedTournament, nil).Once()
			},
			wantErr: nil,
			assertOutput: func(tt *testing.T, output *TournamentOutput) {
				assert.False(tt, output.Active)
			},
		},
		{
			name:         "repo returns ErrTournamentNotFound propagated",
			tournamentID: 99999,
			active:       true,
			setupMock: func(m *MockRepository) {
				m.On("UpdateStatus", mock.Anything, uint(99999), true).Return(ErrTournamentNotFound).Once()
			},
			wantErr: ErrTournamentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			logger := logger.NewLogger("error")
			service := NewService(mockRepo, logger)

			ctx := context.Background()
			tt.setupMock(mockRepo)

			output, err := service.UpdateTournamentStatus(ctx, tt.tournamentID, tt.active)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, output)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				if tt.assertOutput != nil {
					tt.assertOutput(t, output)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestToTournamentOutput(t *testing.T) {
	division := 1
	season := "2024-2025"
	round := 15

	tests := []struct {
		name   string
		input  *Tournament
		assert func(*testing.T, *TournamentOutput)
	}{
		{
			name: "maps all fields correctly",
			input: &Tournament{
				ID:        123,
				Name:      "Premier League",
				Country:   "England",
				Division:  &division,
				Season:    &season,
				Round:     &round,
				Active:    true,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			assert: func(tt *testing.T, output *TournamentOutput) {
				assert.Equal(tt, uint(123), output.ID)
				assert.Equal(tt, "Premier League", output.Name)
				assert.Equal(tt, "England", output.Country)
				assert.Equal(tt, 1, *output.Division)
				assert.Equal(tt, "2024-2025", *output.Season)
				assert.Equal(tt, 15, *output.Round)
				assert.True(tt, output.Active)
				assert.Equal(tt, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), output.CreatedAt)
				assert.Equal(tt, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), output.UpdatedAt)
			},
		},
		{
			name: "handles nil optional fields",
			input: &Tournament{
				ID:        123,
				Name:      "La Liga",
				Country:   "Spain",
				Division:  nil,
				Season:    nil,
				Round:     nil,
				Active:    true,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			assert: func(tt *testing.T, output *TournamentOutput) {
				assert.Equal(tt, uint(123), output.ID)
				assert.Equal(tt, "La Liga", output.Name)
				assert.Equal(tt, "Spain", output.Country)
				assert.Nil(tt, output.Division)
				assert.Nil(tt, output.Season)
				assert.Nil(tt, output.Round)
				assert.True(tt, output.Active)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := toTournamentOutput(tt.input)
			tt.assert(t, output)
		})
	}
}
