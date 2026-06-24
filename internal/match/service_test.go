package match

import (
	"context"
	"errors"
	"testing"

	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

type mockRepository struct {
	matches []*MatchEntity
	total   int64
	listErr error
}

func (m *mockRepository) List(ctx context.Context, filter MatchFilter, page int, pageSize int) ([]*MatchEntity, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.matches, m.total, nil
}

func (m *mockRepository) GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, season string, limit int) ([]*MatchEntity, error) {
	return nil, nil
}

func (m *mockRepository) GetHomeStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error) {
	return VenueStatsEntity{}, nil
}

func (m *mockRepository) GetAwayStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error) {
	return VenueStatsEntity{}, nil
}

func (m *mockRepository) GetOverallStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error) {
	return VenueStatsEntity{}, nil
}

func TestService_ListMatches(t *testing.T) {
	ctx := context.Background()
	lg := logger.NewLogger("info")

	homeOdd := 1.85
	awayOdd := 4.20
	drawOdd := 3.40
	bttsOdd := 1.75
	under25Odd := 1.65
	homeName := "Flamengo"
	awayName := "Palmeiras"
	timestamp := int64(1700000000)

	tests := []struct {
		name        string
		matches     []*MatchEntity
		total       int64
		listErr     error
		page        int
		pageSize    int
		expectError bool
		expectCount int
	}{
		{
			name: "success - list matches",
			matches: []*MatchEntity{
				{ID: 1, LeagueID: 10, Season: "2024", HomeTeamID: 100, AwayTeamID: 200, HomeTeamOdd: &homeOdd, AwayTeamOdd: &awayOdd, DrawOdd: &drawOdd, BTTSOdd: &bttsOdd, Under25Odd: &under25Odd, HomeTeamName: &homeName, AwayTeamName: &awayName, DateTimestamp: &timestamp},
			},
			total:       1,
			page:        1,
			pageSize:    20,
			expectError: false,
			expectCount: 1,
		},
		{
			name: "success - normalize page",
			matches: []*MatchEntity{
				{ID: 1},
			},
			total:       1,
			page:        0,
			pageSize:    20,
			expectError: false,
			expectCount: 1,
		},
		{
			name: "success - normalize page size below minimum",
			matches: []*MatchEntity{
				{ID: 1},
			},
			total:       1,
			page:        1,
			pageSize:    0,
			expectError: false,
			expectCount: 1,
		},
		{
			name: "success - cap page size at max",
			matches: []*MatchEntity{
				{ID: 1},
			},
			total:       1,
			page:        1,
			pageSize:    150,
			expectError: false,
			expectCount: 1,
		},
		{
			name: "success - empty results",
			matches:     []*MatchEntity{},
			total:       0,
			page:        1,
			pageSize:    20,
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "error - repository error",
			matches:     nil,
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
				matches: tt.matches,
				total:   tt.total,
				listErr: tt.listErr,
			}

			svc := NewService(repo, lg)
			outputs, total, err := svc.ListMatches(ctx, MatchFilter{}, tt.page, tt.pageSize)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
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

			if len(outputs) != tt.expectCount {
				t.Errorf("expected %d outputs, got %d", tt.expectCount, len(outputs))
			}
		})
	}
}

func TestService_ListMapsEntityToOutput(t *testing.T) {
	ctx := context.Background()
	lg := logger.NewLogger("info")

	homeOdd := 1.85
	awayOdd := 4.20
	drawOdd := 3.40
	bttsOdd := 1.75
	under25Odd := 1.65
	homeName := "Flamengo"
	awayName := "Palmeiras"
	timestamp := int64(1700000000)

	repo := &mockRepository{
		matches: []*MatchEntity{
			{
				ID:             1,
				LeagueID:       10,
				Season:         "2024",
				Round:          15,
				DateTimestamp:  &timestamp,
				Status:         "finished",
				Time:           90,
				HomeTeamID:     100,
				HomeTeamName:   &homeName,
				HomeTeamGoals:  2,
				HomeTeamOdd:    &homeOdd,
				AwayTeamID:     200,
				AwayTeamName:   &awayName,
				AwayTeamGoals:  1,
				AwayTeamOdd:    &awayOdd,
				DrawOdd:        &drawOdd,
				BTTSOdd:        &bttsOdd,
				Under25Odd:     &under25Odd,
			},
		},
		total: 1,
	}

	svc := NewService(repo, lg)
	outputs, _, err := svc.ListMatches(ctx, MatchFilter{}, 1, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(outputs))
	}

	out := outputs[0]
	if out.ID != 1 {
		t.Errorf("expected ID 1, got %d", out.ID)
	}
	if out.TournamentID != 10 {
		t.Errorf("expected TournamentID 10, got %d", out.TournamentID)
	}
	if out.Season != "2024" {
		t.Errorf("expected Season 2024, got %s", out.Season)
	}
	if out.HomeTeamName == nil || *out.HomeTeamName != "Flamengo" {
		t.Error("expected HomeTeamName Flamengo")
	}
	if out.BTTSOdd == nil || *out.BTTSOdd != 1.75 {
		t.Error("expected BTTSOdd 1.75")
	}
	if out.Under25Odd == nil || *out.Under25Odd != 1.65 {
		t.Error("expected Under25Odd 1.65")
	}
}

func TestService_ListMatchesNullOddsHandling(t *testing.T) {
	ctx := context.Background()
	lg := logger.NewLogger("info")

	repo := &mockRepository{
		matches: []*MatchEntity{
			{ID: 1, HomeTeamOdd: nil, AwayTeamOdd: nil, DrawOdd: nil, BTTSOdd: nil, Under25Odd: nil},
		},
		total: 1,
	}

	svc := NewService(repo, lg)
	outputs, _, err := svc.ListMatches(ctx, MatchFilter{}, 1, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(outputs))
	}

	out := outputs[0]
	if out.HomeTeamOdd != nil {
		t.Error("expected nil HomeTeamOdd")
	}
	if out.BTTSOdd != nil {
		t.Error("expected nil BTTSOdd")
	}
	if out.Under25Odd != nil {
		t.Error("expected nil Under25Odd")
	}
}
