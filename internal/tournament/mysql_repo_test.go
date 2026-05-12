package tournament

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSuite(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.AutoMigrate(&Tournament{})

	return db
}

func TestMysqlTournamentRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		division := 1
		season := "2024-2025"
		round := 15
		active := true

		tournament := &Tournament{
			Name:     "Premier League",
			Country:  "England",
			Division: &division,
			Season:   &season,
			Round:    &round,
			Active:   active,
		}

		err := repo.Create(ctx, tournament)

		assert.NoError(t, err)
		assert.NotZero(t, tournament.ID)
		assert.Equal(t, "Premier League", tournament.Name)
		assert.Equal(t, "England", tournament.Country)
		assert.Equal(t, 1, *tournament.Division)
		assert.Equal(t, "2024-2025", *tournament.Season)
		assert.Equal(t, 15, *tournament.Round)
		assert.True(t, tournament.Active)
	})

	t.Run("success with optional fields nil", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		tournament := &Tournament{
			Name:    "La Liga",
			Country: "Spain",
			Active:  true,
		}

		err := repo.Create(ctx, tournament)

		assert.NoError(t, err)
		assert.NotZero(t, tournament.ID)
		assert.Nil(t, tournament.Division)
		assert.Nil(t, tournament.Season)
		assert.Nil(t, tournament.Round)
	})
}

func TestMysqlTournamentRepository_List(t *testing.T) {
	t.Run("success with multiple rows", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		division1 := 1
		division2 := 2
		season1 := "2024-2025"
		season2 := "2024-2025"

		tournament1 := &Tournament{
			Name:     "Premier League",
			Country:  "England",
			Division: &division1,
			Season:   &season1,
			Active:   true,
		}
		tournament2 := &Tournament{
			Name:     "Championship",
			Country:  "England",
			Division: &division2,
			Season:   &season2,
			Active:   true,
		}
		require.NoError(t, repo.Create(ctx, tournament1))
		require.NoError(t, repo.Create(ctx, tournament2))

		tournaments, total, err := repo.List(ctx, TournamentFilter{}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, tournaments, 2)
		assert.Equal(t, "Championship", tournaments[0].Name)
		assert.Equal(t, "Premier League", tournaments[1].Name)
	})

	t.Run("empty result when no tournaments", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		tournaments, total, err := repo.List(ctx, TournamentFilter{}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, tournaments, 0)
	})

	t.Run("filter by active true", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		season1 := "2024-2025"
		season2 := "2024-2025"

		active := true
		inactive := false

		tournament1 := &Tournament{
			Name:    "Active Tournament",
			Country: "England",
			Season:  &season1,
			Active:  active,
		}
		tournament2 := &Tournament{
			Name:    "Inactive Tournament",
			Country: "England",
			Season:  &season2,
			Active:  inactive,
		}
		require.NoError(t, repo.Create(ctx, tournament1))
		require.NoError(t, repo.Create(ctx, tournament2))

		activeFilter := true
		tournaments, total, err := repo.List(ctx, TournamentFilter{Active: &activeFilter}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, tournaments, 1)
		assert.Equal(t, "Active Tournament", tournaments[0].Name)
	})

	t.Run("filter by active false", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		season1 := "2024-2025"
		season2 := "2024-2025"

		active := true
		inactive := false

		tournament1 := &Tournament{
			Name:    "Active Tournament",
			Country: "England",
			Season:  &season1,
			Active:  active,
		}
		tournament2 := &Tournament{
			Name:    "Inactive Tournament",
			Country: "England",
			Season:  &season2,
			Active:  inactive,
		}
		require.NoError(t, repo.Create(ctx, tournament1))
		require.NoError(t, repo.Create(ctx, tournament2))

		activeFilter := false
		tournaments, total, err := repo.List(ctx, TournamentFilter{Active: &activeFilter}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, tournaments, 1)
		assert.Equal(t, "Inactive Tournament", tournaments[0].Name)
	})

	t.Run("filter by country", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		season1 := "2024-2025"
		season2 := "2024-2025"

		tournament1 := &Tournament{
			Name:    "Premier League",
			Country: "England",
			Season:  &season1,
			Active:  true,
		}
		tournament2 := &Tournament{
			Name:    "La Liga",
			Country: "Spain",
			Season:  &season2,
			Active:  true,
		}
		require.NoError(t, repo.Create(ctx, tournament1))
		require.NoError(t, repo.Create(ctx, tournament2))

		country := "England"
		tournaments, total, err := repo.List(ctx, TournamentFilter{Country: &country}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, tournaments, 1)
		assert.Equal(t, "Premier League", tournaments[0].Name)
	})

	t.Run("filter by division", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		division1 := 1
		division2 := 2
		season1 := "2024-2025"
		season2 := "2024-2025"

		tournament1 := &Tournament{
			Name:     "Premier League",
			Country:  "England",
			Division: &division1,
			Season:   &season1,
			Active:   true,
		}
		tournament2 := &Tournament{
			Name:     "Championship",
			Country:  "England",
			Division: &division2,
			Season:   &season2,
			Active:   true,
		}
		require.NoError(t, repo.Create(ctx, tournament1))
		require.NoError(t, repo.Create(ctx, tournament2))

		division := 1
		tournaments, total, err := repo.List(ctx, TournamentFilter{Division: &division}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, tournaments, 1)
		assert.Equal(t, "Premier League", tournaments[0].Name)
	})

	t.Run("filter by season", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		season1 := "2024-2025"
		season2 := "2023-2024"

		tournament1 := &Tournament{
			Name:    "Premier League 24/25",
			Country: "England",
			Season:  &season1,
			Active:  true,
		}
		tournament2 := &Tournament{
			Name:    "Premier League 23/24",
			Country: "England",
			Season:  &season2,
			Active:  false,
		}
		require.NoError(t, repo.Create(ctx, tournament1))
		require.NoError(t, repo.Create(ctx, tournament2))

		season := "2024-2025"
		tournaments, total, err := repo.List(ctx, TournamentFilter{Season: &season}, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, tournaments, 1)
		assert.Equal(t, "Premier League 24/25", tournaments[0].Name)
	})

	t.Run("pagination works correctly", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		season1 := "2024-2025"
		season2 := "2024-2025"
		season3 := "2024-2025"

		tournament1 := &Tournament{
			Name:    "Tournament 1",
			Country: "England",
			Season:  &season1,
			Active:  true,
		}
		tournament2 := &Tournament{
			Name:    "Tournament 2",
			Country: "England",
			Season:  &season2,
			Active:  true,
		}
		tournament3 := &Tournament{
			Name:    "Tournament 3",
			Country: "England",
			Season:  &season3,
			Active:  true,
		}
		require.NoError(t, repo.Create(ctx, tournament1))
		require.NoError(t, repo.Create(ctx, tournament2))
		require.NoError(t, repo.Create(ctx, tournament3))

		tournaments, total, err := repo.List(ctx, TournamentFilter{}, 1, 2)

		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, tournaments, 2)
	})

	t.Run("out-of-bounds page returns empty slice with correct total", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		season := "2024-2025"

		tournament := &Tournament{
			Name:    "Test Tournament",
			Country: "England",
			Season:  &season,
			Active:  true,
		}
		require.NoError(t, repo.Create(ctx, tournament))

		tournaments, total, err := repo.List(ctx, TournamentFilter{}, 9999, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, tournaments, 0)
	})
}

func TestMysqlTournamentRepository_FindByID(t *testing.T) {
	t.Run("success returns correct tournament", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		division := 1
		season := "2024-2025"
		round := 15

		tournament := &Tournament{
			Name:     "Premier League",
			Country:  "England",
			Division: &division,
			Season:   &season,
			Round:    &round,
			Active:   true,
		}
		require.NoError(t, repo.Create(ctx, tournament))

		found, err := repo.FindByID(ctx, tournament.ID)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, tournament.ID, found.ID)
		assert.Equal(t, "Premier League", found.Name)
		assert.Equal(t, "England", found.Country)
		assert.Equal(t, 1, *found.Division)
		assert.Equal(t, "2024-2025", *found.Season)
		assert.Equal(t, 15, *found.Round)
		assert.True(t, found.Active)
	})

	t.Run("not found returns ErrTournamentNotFound", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		found, err := repo.FindByID(ctx, 99999)

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, ErrTournamentNotFound)
	})
}

func TestMysqlTournamentRepository_Update(t *testing.T) {
	t.Run("success all fields updated updated_at refreshed", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		division := 1
		season := "2024-2025"
		round := 15

		tournament := &Tournament{
			Name:     "Original Name",
			Country:  "England",
			Division: &division,
			Season:   &season,
			Round:    &round,
			Active:   true,
		}
		require.NoError(t, repo.Create(ctx, tournament))

		originalUpdatedAt := tournament.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		newDivision := 2
		newSeason := "2023-2024"
		newRound := 20

		tournament.Name = "Updated Name"
		tournament.Country = "Spain"
		tournament.Division = &newDivision
		tournament.Season = &newSeason
		tournament.Round = &newRound

		err := repo.Update(ctx, tournament)

		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, tournament.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "Spain", found.Country)
		assert.Equal(t, 2, *found.Division)
		assert.Equal(t, "2023-2024", *found.Season)
		assert.Equal(t, 20, *found.Round)
		assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("not found returns ErrTournamentNotFound", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		tournament := &Tournament{
			ID:      99999,
			Name:    "Test Tournament",
			Country: "England",
			Active:  true,
		}

		err := repo.Update(ctx, tournament)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrTournamentNotFound)
	})
}

func TestMysqlTournamentRepository_UpdateStatus(t *testing.T) {
	t.Run("activate sets active=true and updates updated_at", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		tournament := &Tournament{
			Name:    "Test Tournament",
			Country: "England",
			Active:  false,
		}
		require.NoError(t, repo.Create(ctx, tournament))

		originalUpdatedAt := tournament.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		err := repo.UpdateStatus(ctx, tournament.ID, true)

		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, tournament.ID)
		require.NoError(t, err)
		assert.True(t, found.Active)
		assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("deactivate sets active=false and updates updated_at", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		tournament := &Tournament{
			Name:    "Test Tournament",
			Country: "England",
			Active:  true,
		}
		require.NoError(t, repo.Create(ctx, tournament))

		originalUpdatedAt := tournament.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		err := repo.UpdateStatus(ctx, tournament.ID, false)

		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, tournament.ID)
		require.NoError(t, err)
		assert.False(t, found.Active)
		assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("not found returns ErrTournamentNotFound", func(t *testing.T) {
		db := setupSuite(t)
		repo := NewMysqlRepository(db)
		ctx := context.Background()

		err := repo.UpdateStatus(ctx, 99999, true)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrTournamentNotFound)
	})
}
