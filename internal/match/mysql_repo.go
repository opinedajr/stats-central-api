package match

import (
	"context"

	"gorm.io/gorm"
)

type mysqlRepository struct {
	db *gorm.DB
}

func NewMysqlRepository(db *gorm.DB) MatchRepository {
	return &mysqlRepository{
		db: db,
	}
}

func (r *mysqlRepository) GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, limit int) ([]*MatchEntity, error) {
	var matches []*MatchEntity

	err := r.db.WithContext(ctx).
		Table("jogos").
		Select("id, sofascore_id, liga_id, temporada, rodada, data_timestamp, status, tempo, time_mandante_id, time_mandante_gols, time_visitante_id, time_visitante_gols").
		Where("(time_mandante_id = ? OR time_visitante_id = ?) AND liga_id = ? AND status IN (?, ?)", teamID, teamID, tournamentID, "fulltime", "finished").
		Order("data_timestamp DESC").
		Limit(limit).
		Find(&matches).Error

	if err != nil {
		return nil, WrapError(ErrDatabaseError, err.Error())
	}

	return matches, nil
}

func (r *mysqlRepository) GetHomeStats(ctx context.Context, teamID uint, tournamentID uint) (VenueStatsEntity, error) {
	var stats VenueStatsEntity

	query := `
		SELECT
			COUNT(*) as matches_played,
			SUM(CASE WHEN time_mandante_gols > time_visitante_gols THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN time_mandante_gols = time_visitante_gols THEN 1 ELSE 0 END) as draws,
			SUM(CASE WHEN time_mandante_gols < time_visitante_gols THEN 1 ELSE 0 END) as losses,
			COALESCE(SUM(time_mandante_gols), 0) as goals_for,
			COALESCE(SUM(time_visitante_gols), 0) as goals_against
		FROM jogos
		WHERE time_mandante_id = ? AND liga_id = ? AND status IN (?, ?)
	`

	err := r.db.WithContext(ctx).Raw(query, teamID, tournamentID, "fulltime", "finished").Scan(&stats).Error
	if err != nil {
		return VenueStatsEntity{}, WrapError(ErrDatabaseError, err.Error())
	}

	return stats, nil
}

func (r *mysqlRepository) GetAwayStats(ctx context.Context, teamID uint, tournamentID uint) (VenueStatsEntity, error) {
	var stats VenueStatsEntity

	query := `
		SELECT
			COUNT(*) as matches_played,
			SUM(CASE WHEN time_visitante_gols > time_mandante_gols THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN time_visitante_gols = time_mandante_gols THEN 1 ELSE 0 END) as draws,
			SUM(CASE WHEN time_visitante_gols < time_mandante_gols THEN 1 ELSE 0 END) as losses,
			COALESCE(SUM(time_visitante_gols), 0) as goals_for,
			COALESCE(SUM(time_mandante_gols), 0) as goals_against
		FROM jogos
		WHERE time_visitante_id = ? AND liga_id = ? AND status IN (?, ?)
	`

	err := r.db.WithContext(ctx).Raw(query, teamID, tournamentID, "fulltime", "finished").Scan(&stats).Error
	if err != nil {
		return VenueStatsEntity{}, WrapError(ErrDatabaseError, err.Error())
	}

	return stats, nil
}

func (r *mysqlRepository) GetOverallStats(ctx context.Context, teamID uint, tournamentID uint) (VenueStatsEntity, error) {
	var stats VenueStatsEntity

	query := `
		SELECT
			COUNT(*) as matches_played,
			SUM(CASE
				WHEN time_mandante_id = ? AND time_mandante_gols > time_visitante_gols THEN 1
				WHEN time_visitante_id = ? AND time_visitante_gols > time_mandante_gols THEN 1
				ELSE 0
			END) as wins,
			SUM(CASE
				WHEN time_mandante_id = ? AND time_mandante_gols = time_visitante_gols THEN 1
				WHEN time_visitante_id = ? AND time_visitante_gols = time_mandante_gols THEN 1
				ELSE 0
			END) as draws,
			SUM(CASE
				WHEN time_mandante_id = ? AND time_mandante_gols < time_visitante_gols THEN 1
				WHEN time_visitante_id = ? AND time_visitante_gols < time_mandante_gols THEN 1
				ELSE 0
			END) as losses,
			COALESCE(SUM(CASE
				WHEN time_mandante_id = ? THEN time_mandante_gols
				WHEN time_visitante_id = ? THEN time_visitante_gols
				ELSE 0
			END), 0) as goals_for,
			COALESCE(SUM(CASE
				WHEN time_mandante_id = ? THEN time_visitante_gols
				WHEN time_visitante_id = ? THEN time_mandante_gols
				ELSE 0
			END), 0) as goals_against
		FROM jogos
		WHERE (time_mandante_id = ? OR time_visitante_id = ?) AND liga_id = ? AND status IN (?, ?)
	`

	err := r.db.WithContext(ctx).Raw(query, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, tournamentID, "fulltime", "finished").Scan(&stats).Error
	if err != nil {
		return VenueStatsEntity{}, WrapError(ErrDatabaseError, err.Error())
	}

	return stats, nil
}
