package match

type MatchOutput struct {
	ID             uint     `json:"id"`
	TournamentID   uint     `json:"tournament_id"`
	Season         string   `json:"season"`
	Round          int      `json:"round"`
	DateTimestamp  *int64   `json:"date_timestamp"`
	Status         string   `json:"status"`
	Time           int      `json:"time"`
	HomeTeamID     uint     `json:"home_team_id"`
	HomeTeamName   *string  `json:"home_team_name"`
	HomeTeamGoals  int      `json:"home_team_goals"`
	HomeTeamOdd    *float64 `json:"home_team_odd"`
	AwayTeamID     uint     `json:"away_team_id"`
	AwayTeamName   *string  `json:"away_team_name"`
	AwayTeamGoals  int      `json:"away_team_goals"`
	AwayTeamOdd    *float64 `json:"away_team_odd"`
	DrawOdd        *float64 `json:"draw_odd"`
	BTTSOdd        *float64 `json:"btts_odd"`
	Under25Odd     *float64 `json:"under25_odd"`
}
