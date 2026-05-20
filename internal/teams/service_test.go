package teams

import (
	"context"
	"errors"
	"testing"

	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

type mockRepository struct {
	teams      []*Team
	total      int64
	findByIDErr error
	listErr    error
}

func (m *mockRepository) FindByID(ctx context.Context, id uint) (*Team, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	for _, t := range m.teams {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, ErrTeamNotFound
}

func (m *mockRepository) List(ctx context.Context, filter TeamFilter, page int, pageSize int) ([]*Team, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.teams, m.total, nil
}

func TestService_ListTeams(t *testing.T) {
	ctx := context.Background()
	logger := logger.NewLogger("info")

	tests := []struct {
		name        string
		teams       []*Team
		total       int64
		listErr     error
		page        int
		pageSize    int
		expectError bool
	}{
		{
			name: "success - list teams",
			teams: []*Team{
				{ID: 1, Name: "Flamengo", Country: "Brazil"},
				{ID: 2, Name: "Palmeiras", Country: "Brazil"},
			},
			total:       2,
			page:        1,
			pageSize:    20,
			expectError: false,
		},
		{
			name:        "success - normalize page",
			teams:       []*Team{},
			total:       0,
			page:        0,
			pageSize:    20,
			expectError: false,
		},
		{
			name:        "success - normalize page size",
			teams:       []*Team{},
			total:       0,
			page:        1,
			pageSize:    0,
			expectError: false,
		},
		{
			name:        "success - cap page size",
			teams:       []*Team{},
			total:       0,
			page:        1,
			pageSize:    150,
			expectError: false,
		},
		{
			name:        "error - repository error",
			teams:       nil,
			total:       0,
			listErr:     errors.New("database error"),
			page:        1,
			pageSize:    20,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				teams:   tt.teams,
				total:   tt.total,
				listErr: tt.listErr,
			}

			service := NewService(repo, logger)
			outputs, total, err := service.ListTeams(ctx, TeamFilter{}, tt.page, tt.pageSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if total != tt.total {
				t.Errorf("expected total %d, got %d", tt.total, total)
			}

			if len(outputs) != len(tt.teams) {
				t.Errorf("expected %d outputs, got %d", len(tt.teams), len(outputs))
			}
		})
	}
}

func TestService_GetTeamByID(t *testing.T) {
	ctx := context.Background()
	logger := logger.NewLogger("info")

	tests := []struct {
		name        string
		teams       []*Team
		findByID    uint
		findByIDErr error
		expectError bool
	}{
		{
			name: "success - found team",
			teams: []*Team{
				{ID: 1, Name: "Flamengo", Country: "Brazil"},
			},
			findByID:    1,
			findByIDErr: nil,
			expectError: false,
		},
		{
			name:        "error - team not found",
			teams:       []*Team{},
			findByID:    999,
			findByIDErr: ErrTeamNotFound,
			expectError: true,
		},
		{
			name:        "error - repository error",
			teams:       nil,
			findByID:    1,
			findByIDErr: errors.New("database error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				teams:       tt.teams,
				findByIDErr: tt.findByIDErr,
			}

			service := NewService(repo, logger)
			output, err := service.GetTeamByID(ctx, tt.findByID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if output.ID != tt.findByID {
				t.Errorf("expected ID %d, got %d", tt.findByID, output.ID)
			}
		})
	}
}
