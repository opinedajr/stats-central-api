package teams

import "time"

type TeamFilter struct {
	Country *string
	Name    *string
}

type TeamOutput struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Country     string     `json:"country"`
	SofascoreID *int       `json:"sofascore_id,omitempty"`
	SokkerproID *int       `json:"sokkerpro_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
