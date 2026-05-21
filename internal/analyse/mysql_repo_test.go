package analyse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalyseTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE time_estatisticas (
			time_id INTEGER NOT NULL,
			liga_id INTEGER NOT NULL,
			temporada TEXT NOT NULL,
			media_gols_marcados REAL,
			media_gols_sofridos REAL,
			media_cantos_marcados REAL,
			media_cantos_sofridos REAL,
			frequencia_btts REAL,
			frequencia_over05ht REAL,
			frequencia_over15ht REAL,
			frequencia_over15 REAL,
			frequencia_over25 REAL,
			frequencia_over35 REAL,
			frequencia_cantos85 REAL,
			media_gols_marcados_mandante REAL,
			media_gols_sofridos_mandante REAL,
			media_cantos_marcados_mandante REAL,
			media_cantos_sofridos_mandante REAL,
			frequencia_btts_mandante REAL,
			frequencia_over05ht_mandante REAL,
			frequencia_over15ht_mandante REAL,
			frequencia_over15_mandante REAL,
			frequencia_over25_mandante REAL,
			frequencia_over35_mandante REAL,
			frequencia_cantos85_mandante REAL,
			media_gols_marcados_visitante REAL,
			media_gols_sofridos_visitante REAL,
			media_cantos_marcados_visitante REAL,
			media_cantos_sofridos_visitante REAL,
			frequencia_btts_visitante REAL,
			frequencia_over05ht_visitante REAL,
			frequencia_over15ht_visitante REAL,
			frequencia_over15_visitante REAL,
			frequencia_over25_visitante REAL,
			frequencia_over35_visitante REAL,
			frequencia_cantos85_visitante REAL,
			frequencia_primeiro_marcar_mandante REAL,
			frequencia_primeiro_marcar_visitante REAL,
			frequencia_gol_70_mandante REAL,
			frequencia_gol_70_visitante REAL,
			frequencia_gol_45_mandante REAL,
			frequencia_gol_45_visitante REAL,
			frequencia_gol_30_mandante REAL,
			frequencia_gol_30_visitante REAL,
			frequencia_gol_20_mandante REAL,
			frequencia_gol_20_visitante REAL,
			PRIMARY KEY (time_id, liga_id, temporada)
		)
	`).Error
	require.NoError(t, err)

	return db
}

func insertTestStats(t *testing.T, db *gorm.DB, teamID, leagueID int, avgGoalsScored, avgGoalsConceded float64) {
	avgScored := avgGoalsScored
	avgConceded := avgGoalsConceded

	err := db.Exec(`
		INSERT INTO time_estatisticas (
			time_id, liga_id, temporada, media_gols_marcados, media_gols_sofridos,
			media_gols_marcados_mandante, media_gols_sofridos_mandante,
			media_gols_marcados_visitante, media_gols_sofridos_visitante
		) VALUES (?, ?, '2024', ?, ?, ?, ?, ?, ?)
	`, teamID, leagueID, avgScored, avgConceded, avgScored, avgConceded, avgScored, avgConceded).Error
	require.NoError(t, err)
}

func TestStatsRepository_GetTeamStats(t *testing.T) {
	db := setupAnalyseTestDB(t)
	repo := NewMysqlStatsRepository(db)
	ctx := context.Background()

	insertTestStats(t, db, 100, 1, 2.5, 1.2)
	insertTestStats(t, db, 100, 2, 1.8, 0.9)

	t.Run("returns stats for existing team and tournament", func(t *testing.T) {
		stats, err := repo.GetTeamStats(ctx, 100, 1)

		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, uint(100), stats.TeamID)
		assert.Equal(t, uint(1), stats.LeagueID)
		assert.Equal(t, "2024", stats.Season)
		assert.NotNil(t, stats.AvgGoalsScored)
		assert.InDelta(t, 2.5, *stats.AvgGoalsScored, 0.01)
		assert.InDelta(t, 1.2, *stats.AvgGoalsConceded, 0.01)
	})

	t.Run("filters by tournament correctly", func(t *testing.T) {
		stats, err := repo.GetTeamStats(ctx, 100, 2)

		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.InDelta(t, 1.8, *stats.AvgGoalsScored, 0.01)
	})

	t.Run("returns ErrStatsNotFound when no matching record", func(t *testing.T) {
		stats, err := repo.GetTeamStats(ctx, 999, 1)

		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Equal(t, ErrStatsNotFound, err)
	})

	t.Run("handles nil pointer fields correctly", func(t *testing.T) {
		err := db.Exec(`
			INSERT INTO time_estatisticas (time_id, liga_id, temporada)
			VALUES (200, 1, '2024')
		`).Error
		require.NoError(t, err)

		stats, err := repo.GetTeamStats(ctx, 200, 1)

		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Nil(t, stats.AvgGoalsScored)
		assert.Nil(t, stats.AvgGoalsConceded)
	})
}
