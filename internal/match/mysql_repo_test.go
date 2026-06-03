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
			sofascore_id INTEGER,
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
				primeiro_marcar INTEGER,
				segundo_marcar INTEGER,
				terceiro_marcar INTEGER,
				minuto_gol1 INTEGER,
				minuto_gol2 INTEGER,
				minuto_gol3 INTEGER
			)
		`).Error
	require.NoError(t, err)

	return db
}

func insertTestMatch(t *testing.T, db *gorm.DB, id int, leagueID, homeTeamID, homeGoals, awayTeamID, awayGoals int, status string, timestamp int64, season string, tempo int) {
	err := db.Exec(`
		INSERT INTO jogos (id, sofascore_id, liga_id, temporada, rodada, data_timestamp, status, tempo, time_mandante_id, time_mandante_gols, time_visitante_id, time_visitante_gols)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, nil, leagueID, season, 1, timestamp, status, tempo, homeTeamID, homeGoals, awayTeamID, awayGoals).Error
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

