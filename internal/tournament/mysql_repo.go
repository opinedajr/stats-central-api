package tournament

import (
	"context"

	"gorm.io/gorm"
)

type mysqlTournamentRepository struct {
	db *gorm.DB
}

func NewMysqlTournamentRepository(db *gorm.DB) TournamentRepository {
	return &mysqlTournamentRepository{
		db: db,
	}
}

func (r *mysqlTournamentRepository) Create(ctx context.Context, tournament *Tournament) error {
	if err := r.db.WithContext(ctx).Create(tournament).Error; err != nil {
		return WrapError(ErrDatabaseError, err.Error())
	}
	return nil
}

func (r *mysqlTournamentRepository) Update(ctx context.Context, tournament *Tournament) error {
	result := r.db.WithContext(ctx).Model(&Tournament{}).
		Where("id = ?", tournament.ID).
		Updates(map[string]interface{}{
			"name":     tournament.Name,
			"country":  tournament.Country,
			"division": tournament.Division,
			"season":   tournament.Season,
			"round":    tournament.Round,
		})

	if result.Error != nil {
		return WrapError(ErrDatabaseError, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return ErrTournamentNotFound
	}
	return nil
}

func (r *mysqlTournamentRepository) UpdateStatus(ctx context.Context, id uint, active bool) error {
	result := r.db.WithContext(ctx).Model(&Tournament{}).
		Where("id = ?", id).
		Update("active", active)

	if result.Error != nil {
		return WrapError(ErrDatabaseError, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return ErrTournamentNotFound
	}
	return nil
}

func (r *mysqlTournamentRepository) FindByID(ctx context.Context, id uint) (*Tournament, error) {
	var tournament Tournament
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tournament).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTournamentNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return &tournament, nil
}

func (r *mysqlTournamentRepository) List(ctx context.Context, filter TournamentFilter, page int, pageSize int) ([]*Tournament, int64, error) {
	var tournaments []*Tournament
	var total int64

	query := r.db.WithContext(ctx).Model(&Tournament{})

	if filter.Active != nil {
		query = query.Where("active = ?", *filter.Active)
	}
	if filter.Country != nil {
		query = query.Where("country = ?", *filter.Country)
	}
	if filter.Division != nil {
		query = query.Where("division = ?", *filter.Division)
	}
	if filter.Season != nil {
		query = query.Where("season = ?", *filter.Season)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	offset := (page - 1) * pageSize

	err = query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tournaments).Error

	if err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	return tournaments, total, nil
}
