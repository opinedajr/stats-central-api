package match

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE jogos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			liga_id INTEGER NOT NULL,
			temporada TEXT NOT NULL,
			rodada INTEGER NOT NULL,
			data_timestamp INTEGER,
			status TEXT,
			tempo INTEGER NOT NULL,
			time_mandante_id INTEGER NOT NULL,
			time_mandante_gols INTEGER NOT NULL,
			time_visitante_id INTEGER NOT NULL,
			time_visitante_gols INTEGER NOT NULL DEFAULT 0,
			time_mandante_odd REAL,
			time_visitante_odd REAL,
			empate_odd REAL,
			btts_odd REAL,
			under25_odd REAL,
			primeiro_marcar INTEGER,
			segundo_marcar INTEGER,
			terceiro_marcar INTEGER,
			minuto_gol1 INTEGER,
			minuto_gol2 INTEGER,
			minuto_gol3 INTEGER
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE teams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			country TEXT NOT NULL DEFAULT '',
			sofascore_id INTEGER,
			sokkerpro_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

func insertTestMatch(t *testing.T, db *gorm.DB, id int, leagueID, homeTeamID, homeGoals, awayTeamID, awayGoals int, status string, timestamp int64, season string, tempo int) {
	err := db.Exec(`
		INSERT INTO jogos (id, liga_id, temporada, rodada, data_timestamp, status, tempo, time_mandante_id, time_mandante_gols, time_visitante_id, time_visitante_gols)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, leagueID, season, 1, timestamp, status, tempo, homeTeamID, homeGoals, awayTeamID, awayGoals).Error
	require.NoError(t, err)
}

func insertTestTeam(t *testing.T, db *gorm.DB, id int, name string) {
	err := db.Exec(`INSERT INTO teams (id, name, country) VALUES (?, ?, 'Brazil')`, id, name).Error
	require.NoError(t, err)
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	err := db.Exec("DELETE FROM jogos").Error
	require.NoError(t, err)
}

func TestMysqlRepository_GetRecentMatches(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	insertTestMatch(t, db, 1, 1, 100, 2, 200, 1, "fulltime", 1700000000, "2024", 3)
	insertTestMatch(t, db, 2, 1, 100, 1, 201, 1, "finished", 1700001000, "2024", 3)
	insertTestMatch(t, db, 3, 1, 202, 3, 100, 0, "fulltime", 1700002000, "2024", 3)
	insertTestMatch(t, db, 4, 1, 100, 2, 203, 2, "notstarted", 1700003000, "2024", 3)

	t.Run("returns matches for team home and away ordered by timestamp desc", func(t *testing.T) {
		matches, err := repo.GetRecentMatches(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		assert.Len(t, matches, 3)
		assert.Equal(t, uint(3), matches[0].ID)
		assert.Equal(t, uint(2), matches[1].ID)
		assert.Equal(t, uint(1), matches[2].ID)
	})

	t.Run("filters by status fulltime and finished only", func(t *testing.T) {
		matches, err := repo.GetRecentMatches(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		for _, m := range matches {
			assert.True(t, m.Status == "fulltime" || m.Status == "finished")
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		matches, err := repo.GetRecentMatches(ctx, 100, 1, "2024", 2)

		require.NoError(t, err)
		assert.Len(t, matches, 2)
	})
}

func TestMysqlRepository_GetHomeStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	insertTestMatch(t, db, 1, 1, 100, 2, 200, 1, "fulltime", 1700000000, "2024", 3)
	insertTestMatch(t, db, 2, 1, 100, 1, 201, 1, "finished", 1700001000, "2024", 3)
	insertTestMatch(t, db, 3, 1, 100, 0, 202, 2, "fulltime", 1700002000, "2024", 3)
	insertTestMatch(t, db, 4, 2, 100, 3, 300, 1, "fulltime", 1700003000, "2024", 3)

	t.Run("calculates home stats correctly", func(t *testing.T) {
		stats, err := repo.GetHomeStats(ctx, 100, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 3, stats.MatchesPlayed)
		assert.Equal(t, 1, stats.Wins)
		assert.Equal(t, 1, stats.Draws)
		assert.Equal(t, 1, stats.Losses)
		assert.Equal(t, 3, stats.GoalsFor)
		assert.Equal(t, 4, stats.GoalsAgainst)
	})

	t.Run("returns zero values when no matches", func(t *testing.T) {
		stats, err := repo.GetHomeStats(ctx, 999, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 0, stats.MatchesPlayed)
		assert.Equal(t, 0, stats.Wins)
		assert.Equal(t, 0, stats.Draws)
		assert.Equal(t, 0, stats.Losses)
		assert.Equal(t, 0, stats.GoalsFor)
		assert.Equal(t, 0, stats.GoalsAgainst)
	})
}

func TestMysqlRepository_GetAwayStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	insertTestMatch(t, db, 1, 1, 200, 2, 100, 3, "fulltime", 1700000000, "2024", 3)
	insertTestMatch(t, db, 2, 1, 201, 1, 100, 1, "finished", 1700001000, "2024", 3)
	insertTestMatch(t, db, 3, 1, 202, 2, 100, 0, "fulltime", 1700002000, "2024", 3)
	insertTestMatch(t, db, 4, 2, 300, 1, 100, 3, "fulltime", 1700003000, "2024", 3)

	t.Run("calculates away stats correctly", func(t *testing.T) {
		stats, err := repo.GetAwayStats(ctx, 100, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 3, stats.MatchesPlayed)
		assert.Equal(t, 1, stats.Wins)
		assert.Equal(t, 1, stats.Draws)
		assert.Equal(t, 1, stats.Losses)
		assert.Equal(t, 4, stats.GoalsFor)
		assert.Equal(t, 5, stats.GoalsAgainst)
	})

	t.Run("returns zero values when no matches", func(t *testing.T) {
		stats, err := repo.GetAwayStats(ctx, 999, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 0, stats.MatchesPlayed)
		assert.Equal(t, 0, stats.Wins)
		assert.Equal(t, 0, stats.Draws)
		assert.Equal(t, 0, stats.Losses)
		assert.Equal(t, 0, stats.GoalsFor)
		assert.Equal(t, 0, stats.GoalsAgainst)
	})
}

func TestMysqlRepository_GetOverallStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	insertTestMatch(t, db, 1, 1, 100, 2, 200, 1, "fulltime", 1700000000, "2024", 3)
	insertTestMatch(t, db, 2, 1, 100, 1, 201, 1, "finished", 1700001000, "2024", 3)
	insertTestMatch(t, db, 3, 1, 202, 2, 100, 0, "fulltime", 1700002000, "2024", 3)
	insertTestMatch(t, db, 4, 1, 200, 3, 100, 1, "fulltime", 1700003000, "2024", 3)
	insertTestMatch(t, db, 5, 2, 100, 3, 300, 1, "fulltime", 1700004000, "2024", 3)

	t.Run("calculates overall stats correctly combining home and away", func(t *testing.T) {
		stats, err := repo.GetOverallStats(ctx, 100, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 4, stats.MatchesPlayed)
		assert.Equal(t, 1, stats.Wins)
		assert.Equal(t, 1, stats.Draws)
		assert.Equal(t, 2, stats.Losses)
		assert.Equal(t, 4, stats.GoalsFor)
		assert.Equal(t, 7, stats.GoalsAgainst)
	})

	t.Run("filters by tournament", func(t *testing.T) {
		stats, err := repo.GetOverallStats(ctx, 100, 2, "2024")

		require.NoError(t, err)
		assert.Equal(t, 1, stats.MatchesPlayed)
		assert.Equal(t, 1, stats.Wins)
	})

	t.Run("returns zero values when no matches", func(t *testing.T) {
		stats, err := repo.GetOverallStats(ctx, 999, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 0, stats.MatchesPlayed)
		assert.Equal(t, 0, stats.Wins)
		assert.Equal(t, 0, stats.Draws)
		assert.Equal(t, 0, stats.Losses)
		assert.Equal(t, 0, stats.GoalsFor)
		assert.Equal(t, 0, stats.GoalsAgainst)
	})
}

func TestMysqlRepository_edgeCases(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	t.Run("all wins scenario", func(t *testing.T) {
		insertTestMatch(t, db, 1, 1, 100, 3, 200, 0, "fulltime", 1700000000, "2024", 3)
		insertTestMatch(t, db, 2, 1, 201, 0, 100, 2, "fulltime", 1700001000, "2024", 3)

		stats, err := repo.GetOverallStats(ctx, 100, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 2, stats.MatchesPlayed)
		assert.Equal(t, 2, stats.Wins)
		assert.Equal(t, 0, stats.Draws)
		assert.Equal(t, 0, stats.Losses)
	})

	t.Run("all losses scenario", func(t *testing.T) {
		insertTestMatch(t, db, 10, 1, 100, 0, 200, 3, "fulltime", 1700002000, "2024", 3)
		insertTestMatch(t, db, 20, 1, 201, 2, 100, 0, "fulltime", 1700003000, "2024", 3)

		stats, err := repo.GetOverallStats(ctx, 100, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 4, stats.MatchesPlayed)
		assert.Equal(t, 2, stats.Wins)
		assert.Equal(t, 0, stats.Draws)
		assert.Equal(t, 2, stats.Losses)
	})
}

func TestMysqlRepository_filtersBySeason(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	insertTestMatch(t, db, 1, 1, 100, 2, 200, 1, "fulltime", 1700000000, "2024", 3)
	insertTestMatch(t, db, 2, 1, 100, 1, 201, 1, "finished", 1700001000, "2024", 3)
	insertTestMatch(t, db, 3, 1, 100, 3, 202, 0, "fulltime", 1600000000, "2023", 3)
	insertTestMatch(t, db, 4, 1, 100, 1, 203, 1, "finished", 1600001000, "2023", 3)

	t.Run("GetRecentMatches filters by season", func(t *testing.T) {
		matches2024, err := repo.GetRecentMatches(ctx, 100, 1, "2024", 10)

		require.NoError(t, err)
		assert.Len(t, matches2024, 2)

		matches2023, err := repo.GetRecentMatches(ctx, 100, 1, "2023", 10)

		require.NoError(t, err)
		assert.Len(t, matches2023, 2)
	})

	t.Run("GetHomeStats filters by season", func(t *testing.T) {
		stats2024, err := repo.GetHomeStats(ctx, 100, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 2, stats2024.MatchesPlayed)

		stats2023, err := repo.GetHomeStats(ctx, 100, 1, "2023")

		require.NoError(t, err)
		assert.Equal(t, 2, stats2023.MatchesPlayed)
	})

	t.Run("GetOverallStats filters by season", func(t *testing.T) {
		stats2024, err := repo.GetOverallStats(ctx, 100, 1, "2024")

		require.NoError(t, err)
		assert.Equal(t, 2, stats2024.MatchesPlayed)

		stats2023, err := repo.GetOverallStats(ctx, 100, 1, "2023")

		require.NoError(t, err)
		assert.Equal(t, 2, stats2023.MatchesPlayed)
	})
}

func TestMysqlRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	insertTestTeam(t, db, 100, "Flamengo")
	insertTestTeam(t, db, 200, "Palmeiras")
	insertTestTeam(t, db, 201, "Corinthians")
	insertTestTeam(t, db, 202, "São Paulo")
	insertTestTeam(t, db, 203, "Santos")

	insertTestMatch(t, db, 1, 10, 100, 2, 200, 1, "finished", 1700000000, "2024", 3)
	insertTestMatch(t, db, 2, 10, 201, 1, 202, 0, "finished", 1700001000, "2024", 3)
	insertTestMatch(t, db, 3, 10, 100, 0, 203, 2, "fulltime", 1700002000, "2024", 3)
	insertTestMatch(t, db, 4, 20, 100, 1, 200, 1, "finished", 1700003000, "2024", 3)
	insertTestMatch(t, db, 5, 10, 200, 3, 100, 0, "notstarted", 1700004000, "2024", 3)

	t.Run("returns all matches with no filters", func(t *testing.T) {
		matches, total, err := repo.List(ctx, MatchFilter{}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, matches, 5)
	})

	t.Run("orders by data_timestamp DESC", func(t *testing.T) {
		matches, _, err := repo.List(ctx, MatchFilter{}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, uint(5), matches[0].ID)
		assert.Equal(t, uint(4), matches[1].ID)
	})

	t.Run("filters by tournament_id", func(t *testing.T) {
		tournamentID := uint(10)
		matches, total, err := repo.List(ctx, MatchFilter{TournamentID: &tournamentID}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(4), total)
		assert.Len(t, matches, 4)
	})

	t.Run("filters by season", func(t *testing.T) {
		season := "2024"
		matches, total, err := repo.List(ctx, MatchFilter{Season: &season}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, matches, 5)
	})

	t.Run("filters by status", func(t *testing.T) {
		status := "finished"
		matches, total, err := repo.List(ctx, MatchFilter{Status: &status}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, matches, 3)
	})

	t.Run("filters by home_team_id", func(t *testing.T) {
		homeTeamID := uint(100)
		matches, total, err := repo.List(ctx, MatchFilter{HomeTeamID: &homeTeamID}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, matches, 3)
	})

	t.Run("filters by away_team_id", func(t *testing.T) {
		awayTeamID := uint(100)
		matches, total, err := repo.List(ctx, MatchFilter{AwayTeamID: &awayTeamID}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, matches, 1)
	})

	t.Run("combines multiple filters with AND logic", func(t *testing.T) {
		tournamentID := uint(10)
		status := "finished"
		homeTeamID := uint(100)
		matches, total, err := repo.List(ctx, MatchFilter{
			TournamentID: &tournamentID,
			Status:       &status,
			HomeTeamID:   &homeTeamID,
		}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, matches, 1)
		assert.Equal(t, uint(1), matches[0].ID)
	})

	t.Run("returns empty slice when no matches found", func(t *testing.T) {
		tournamentID := uint(999)
		matches, total, err := repo.List(ctx, MatchFilter{TournamentID: &tournamentID}, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, matches)
	})

	t.Run("paginates correctly", func(t *testing.T) {
		matches, total, err := repo.List(ctx, MatchFilter{}, 1, 2)

		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, matches, 2)

		matches2, _, err := repo.List(ctx, MatchFilter{}, 2, 2)

		require.NoError(t, err)
		assert.Len(t, matches2, 2)

		matches3, _, err := repo.List(ctx, MatchFilter{}, 3, 2)

		require.NoError(t, err)
		assert.Len(t, matches3, 1)
	})

	t.Run("populates team names from teams table", func(t *testing.T) {
		matches, _, err := repo.List(ctx, MatchFilter{}, 1, 5)

		require.NoError(t, err)
		assert.Equal(t, "Flamengo", *matches[4].HomeTeamName)
		assert.Equal(t, "Palmeiras", *matches[4].AwayTeamName)
	})

	t.Run("returns page beyond available with empty results", func(t *testing.T) {
		matches, total, err := repo.List(ctx, MatchFilter{}, 100, 20)

		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Empty(t, matches)
	})
}

func TestMysqlRepository_List_filtersByTempo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMysqlRepository(db)
	ctx := context.Background()

	insertTestMatch(t, db, 1, 10, 100, 2, 200, 1, "fulltime", 1700000000, "2024", 1)
	insertTestMatch(t, db, 2, 10, 100, 2, 200, 1, "fulltime", 1700000000, "2024", 2)
	insertTestMatch(t, db, 3, 10, 100, 2, 200, 1, "fulltime", 1700000000, "2024", 3)

	matches, total, err := repo.List(ctx, MatchFilter{}, 1, 20)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, matches, 1)
	assert.Equal(t, uint(3), matches[0].ID)
	assert.Equal(t, FullTimeMatchDuration, matches[0].Time)
}
