package tournament

import "time"

type Tournament struct {
	ID                uint      `gorm:"primaryKey;autoIncrement"`
	SokkerproID       *int      `gorm:"column:sokkerpro_id"`
	PlayscoresID      *int      `gorm:"column:playscores_id"`
	SofascoreID       *int      `gorm:"column:sofascore_id"`
	SofascoreSeasonID *int      `gorm:"column:sofascore_season_id"`
	Active            bool      `gorm:"column:active;not null"`
	Name              string    `gorm:"column:name;type:varchar(40);not null"`
	Country           string    `gorm:"column:country;type:varchar(40);not null"`
	Division          *int      `gorm:"column:division"`
	Season            *string   `gorm:"column:season;type:varchar(20)"`
	Round             *int      `gorm:"column:round"`
	Stats             bool      `gorm:"column:stats;not null"`
	BotGols           bool      `gorm:"column:bot_gols;default:false"`
	BotCantos         bool      `gorm:"column:bot_cantos;default:false"`
	BotMinutos        bool      `gorm:"column:bot_minutos;default:false"`
	BotRajada         bool      `gorm:"column:bot_rajada;default:false"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Tournament) TableName() string {
	return "tournaments"
}
