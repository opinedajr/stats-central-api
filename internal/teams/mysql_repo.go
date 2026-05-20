package teams

import (
	"context"
	"errors"

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

func (r *mysqlRepository) FindByID(ctx context.Context, id uint) (*Team, error) {
	var team Team
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return &team, nil
}

func (r *mysqlRepository) List(ctx context.Context, filter TeamFilter, page int, pageSize int) ([]*Team, int64, error) {
	var teams []*Team
	var total int64

	query := r.db.WithContext(ctx).Model(&Team{})

	if filter.Country != nil {
		query = query.Where("LOWER(country) = LOWER(?)", *filter.Country)
	}
	if filter.Name != nil {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+*filter.Name+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	offset := (page - 1) * pageSize

	err = query.
		Order("name ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&teams).Error

	if err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	return teams, total, nil
}
