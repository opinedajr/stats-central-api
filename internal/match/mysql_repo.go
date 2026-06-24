package match

import (
	"context"

	"gorm.io/gorm"
)

type mysqlRepository struct {
	db *gorm.DB
}

func NewMysqlRepository(db *gorm.DB) Repository {
	return &mysqlRepository{
		db: db,
	}
}

func (r *mysqlRepository) List(ctx context.Context, filter MatchFilter, page int, pageSize int) ([]*MatchEntity, int64, error) {
	var matches []*MatchEntity
	var total int64

	query := r.db.WithContext(ctx).Table("jogos").
		Where("tempo = ?", FullTimeMatchDuration)

	if filter.TournamentID != nil {
		query = query.Where("liga_id = ?", *filter.TournamentID)
	}
	if filter.Season != nil {
		query = query.Where("temporada = ?", *filter.Season)
	}
	if filter.Round != nil {
		query = query.Where("rodada = ?", *filter.Round)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.HomeTeamID != nil {
		query = query.Where("time_mandante_id = ?", *filter.HomeTeamID)
	}
	if filter.AwayTeamID != nil {
		query = query.Where("time_visitante_id = ?", *filter.AwayTeamID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	offset := (page - 1) * pageSize
	if err := query.
		Order("data_timestamp DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&matches).Error; err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	if len(matches) == 0 {
		return matches, total, nil
	}

	teamIDs := make(map[uint]struct{})
	for _, m := range matches {
		teamIDs[m.HomeTeamID] = struct{}{}
		teamIDs[m.AwayTeamID] = struct{}{}
	}

	ids := make([]uint, 0, len(teamIDs))
	for id := range teamIDs {
		ids = append(ids, id)
	}

	type teamName struct {
		ID   uint   `gorm:"primaryKey;column:id"`
		Name string `gorm:"column:name"`
	}
	var teamNames []teamName
	if err := r.db.WithContext(ctx).Table("teams").
		Select("id, name").
		Where("id IN ?", ids).
		Find(&teamNames).Error; err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	nameMap := make(map[uint]string, len(teamNames))
	for _, t := range teamNames {
		nameMap[t.ID] = t.Name
	}

	for _, m := range matches {
		if name, ok := nameMap[m.HomeTeamID]; ok {
			m.HomeTeamName = &name
		}
		if name, ok := nameMap[m.AwayTeamID]; ok {
			m.AwayTeamName = &name
		}
	}

	return matches, total, nil
}

func (r *mysqlRepository) GetRecentMatches(ctx context.Context, teamID uint, tournamentID uint, season string, limit int) ([]*MatchEntity, error) {
	var matches []*MatchEntity

	err := r.db.WithContext(ctx).
		Table("jogos").
		Select("id, liga_id, temporada, rodada, data_timestamp, status, tempo, time_mandante_id, time_mandante_gols, time_mandante_odd, time_visitante_id, time_visitante_gols, time_visitante_odd, empate_odd, btts_odd, under25_odd, primeiro_marcar, segundo_marcar, terceiro_marcar, minuto_gol1, minuto_gol2, minuto_gol3").
		Where("(time_mandante_id = ? OR time_visitante_id = ?) AND liga_id = ? AND temporada = ? AND status IN (?, ?) AND tempo = ?", teamID, teamID, tournamentID, season, StatusFulltime, StatusFinished, FullTimeMatchDuration).
		Order("data_timestamp DESC").
		Limit(limit).
		Find(&matches).Error

	if err != nil {
		return nil, WrapError(ErrDatabaseError, err.Error())
	}

	return matches, nil
}

func (r *mysqlRepository) GetHomeStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error) {
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
			WHERE time_mandante_id = ? AND liga_id = ? AND temporada = ? AND status IN (?, ?) AND tempo = ?
		`

	err := r.db.WithContext(ctx).Raw(query, teamID, tournamentID, season, StatusFulltime, StatusFinished, FullTimeMatchDuration).Scan(&stats).Error
	if err != nil {
		return VenueStatsEntity{}, WrapError(ErrDatabaseError, err.Error())
	}

	return stats, nil
}

func (r *mysqlRepository) GetAwayStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error) {
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
			WHERE time_visitante_id = ? AND liga_id = ? AND temporada = ? AND status IN (?, ?) AND tempo = ?
		`

	err := r.db.WithContext(ctx).Raw(query, teamID, tournamentID, season, StatusFulltime, StatusFinished, FullTimeMatchDuration).Scan(&stats).Error
	if err != nil {
		return VenueStatsEntity{}, WrapError(ErrDatabaseError, err.Error())
	}

	return stats, nil
}

func (r *mysqlRepository) GetOverallStats(ctx context.Context, teamID uint, tournamentID uint, season string) (VenueStatsEntity, error) {
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
			WHERE (time_mandante_id = ? OR time_visitante_id = ?) AND liga_id = ? AND temporada = ? AND status IN (?, ?) AND tempo = ?
		`

	err := r.db.WithContext(ctx).Raw(query, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, teamID, tournamentID, season, StatusFulltime, StatusFinished, FullTimeMatchDuration).Scan(&stats).Error
	if err != nil {
		return VenueStatsEntity{}, WrapError(ErrDatabaseError, err.Error())
	}

	return stats, nil
}
