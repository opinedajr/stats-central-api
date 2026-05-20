package teams

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&Team{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func insertTestTeam(t *testing.T, db *gorm.DB, team *Team) {
	if err := db.Create(team).Error; err != nil {
		t.Fatalf("failed to insert test team: %v", err)
	}
}

func TestMysqlRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	team1 := &Team{ID: 1, Name: "Flamengo", Country: "Brazil"}
	team2 := &Team{ID: 2, Name: "Palmeiras", Country: "Brazil"}
	insertTestTeam(t, db, team1)
	insertTestTeam(t, db, team2)

	tests := []struct {
		name        string
		id          uint
		expectError bool
		expectedID  uint
	}{
		{
			name:        "success - found team",
			id:          1,
			expectError: false,
			expectedID:  1,
		},
		{
			name:        "error - team not found",
			id:          999,
			expectError: true,
			expectedID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team, err := repo.FindByID(ctx, tt.id)

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

			if team.ID != tt.expectedID {
				t.Errorf("expected ID %d, got %d", tt.expectedID, team.ID)
			}
		})
	}
}

func TestMysqlRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	team1 := &Team{ID: 1, Name: "Flamengo", Country: "Brazil"}
	team2 := &Team{ID: 2, Name: "Palmeiras", Country: "Brazil"}
	team3 := &Team{ID: 3, Name: "Real Madrid", Country: "Spain"}
	insertTestTeam(t, db, team1)
	insertTestTeam(t, db, team2)
	insertTestTeam(t, db, team3)

	tests := []struct {
		name          string
		filter        TeamFilter
		page          int
		pageSize      int
		expectCount   int
		expectTotal   int64
	}{
		{
			name:        "success - list all",
			filter:      TeamFilter{},
			page:        1,
			pageSize:    20,
			expectCount: 3,
			expectTotal: 3,
		},
		{
			name: "success - filter by country",
			filter: TeamFilter{
				Country: stringPtr("Brazil"),
			},
			page:        1,
			pageSize:    20,
			expectCount: 2,
			expectTotal: 2,
		},
		{
			name: "success - filter by name partial",
			filter: TeamFilter{
				Name: stringPtr("Fla"),
			},
			page:        1,
			pageSize:    20,
			expectCount: 1,
			expectTotal: 1,
		},
		{
			name: "success - filter by name case insensitive",
			filter: TeamFilter{
				Name: stringPtr("fla"),
			},
			page:        1,
			pageSize:    20,
			expectCount: 1,
			expectTotal: 1,
		},
		{
			name: "success - combined filters",
			filter: TeamFilter{
				Country: stringPtr("Brazil"),
				Name:    stringPtr("Flamengo"),
			},
			page:        1,
			pageSize:    20,
			expectCount: 1,
			expectTotal: 1,
		},
		{
			name:        "success - pagination",
			filter:      TeamFilter{},
			page:        1,
			pageSize:    2,
			expectCount: 2,
			expectTotal: 3,
		},
		{
			name: "success - filter with no results",
			filter: TeamFilter{
				Country: stringPtr("Argentina"),
			},
			page:        1,
			pageSize:    20,
			expectCount: 0,
			expectTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams, total, err := repo.List(ctx, tt.filter, tt.page, tt.pageSize)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(teams) != tt.expectCount {
				t.Errorf("expected %d teams, got %d", tt.expectCount, len(teams))
			}

			if total != tt.expectTotal {
				t.Errorf("expected total %d, got %d", tt.expectTotal, total)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
