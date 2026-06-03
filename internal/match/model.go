package match

type MatchEntity struct {
	ID             uint  `gorm:"primaryKey;column:id"`
	SofascoreID    *int  `gorm:"column:sofascore_id"`
	LeagueID       uint  `gorm:"column:liga_id"`
	Season         string `gorm:"column:temporada"`
	Round          int    `gorm:"column:rodada"`
	DateTimestamp  *int64 `gorm:"column:data_timestamp"`
	Status         string `gorm:"column:status"`
	Time           int    `gorm:"column:tempo"`
	HomeTeamID     uint   `gorm:"column:time_mandante_id"`
	HomeTeamGoals  int    `gorm:"column:time_mandante_gols"`
	HomeTeamOdd    *float64 `gorm:"column:time_mandante_odd"`
	AwayTeamID     uint   `gorm:"column:time_visitante_id"`
	AwayTeamGoals  int    `gorm:"column:time_visitante_gols"`
	AwayTeamOdd    *float64 `gorm:"column:time_visitante_odd"`
	DrawOdd        *float64 `gorm:"column:empate_odd"`
	FirstToScore   *int      `gorm:"column:primeiro_marcar"`
	SecondToScore  *int      `gorm:"column:segundo_marcar"`
	ThirdToScore   *int      `gorm:"column:terceiro_marcar"`
	Goal1Minute    *int      `gorm:"column:minuto_gol1"`
	Goal2Minute    *int      `gorm:"column:minuto_gol2"`
	Goal3Minute    *int      `gorm:"column:minuto_gol3"`
}

func (MatchEntity) TableName() string {
	return "jogos"
}

type VenueStatsEntity struct {
	MatchesPlayed int
	Wins          int
	Draws         int
	Losses        int
	GoalsFor      int
	GoalsAgainst  int
}
