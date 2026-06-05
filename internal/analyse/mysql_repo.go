package analyse

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type mysqlStatsRepository struct {
	db *gorm.DB
}

func NewMysqlStatsRepository(db *gorm.DB) StatsRepository {
	return &mysqlStatsRepository{
		db: db,
	}
}

func (r *mysqlStatsRepository) GetTeamStats(ctx context.Context, teamID uint, tournamentID uint, season string) (*TeamStatsEntity, error) {
	var stats TeamStatsEntity

	err := r.db.WithContext(ctx).
		Where("time_id = ? AND liga_id = ? AND temporada = ?", teamID, tournamentID, season).
		First(&stats).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStatsNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}

	return &stats, nil
}
