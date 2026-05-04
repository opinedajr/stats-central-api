package tournament

import "time"

type CreateTournamentInput struct {
	Name     string  `json:"name" binding:"required,min=1,max=40"`
	Country  string  `json:"country" binding:"required,min=1,max=40"`
	Division *int    `json:"division"`
	Season   *string `json:"season" binding:"omitempty,max=20"`
	Round    *int    `json:"round"`
	Active   *bool   `json:"active"`
}

type UpdateTournamentInput struct {
	Name     string  `json:"name" binding:"required,min=1,max=40"`
	Country  string  `json:"country" binding:"required,min=1,max=40"`
	Division *int    `json:"division"`
	Season   *string `json:"season" binding:"omitempty,max=20"`
	Round    *int    `json:"round"`
}

type UpdateTournamentStatusInput struct {
	Active bool `json:"active"`
}

type TournamentFilter struct {
	Active   *bool
	Country  *string
	Division *int
	Season   *string
}

type TournamentOutput struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Country   string    `json:"country"`
	Division  *int      `json:"division"`
	Season    *string   `json:"season"`
	Round     *int      `json:"round"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
