package teams

import "time"

type Team struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	SofascoreID *int      `gorm:"column:sofascore_id"`
	SokkerproID *int      `gorm:"column:sokkerpro_id"`
	Name        string    `gorm:"column:name;type:varchar(40);not null"`
	Country     string    `gorm:"column:country;type:varchar(40);not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Team) TableName() string {
	return "teams"
}
