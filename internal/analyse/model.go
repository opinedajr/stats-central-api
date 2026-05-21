package analyse

import "time"

type TeamStatsEntity struct {
	TeamID   uint   `gorm:"column:time_id;primaryKey"`
	LeagueID uint   `gorm:"column:liga_id;primaryKey"`
	Season   string `gorm:"column:temporada;primaryKey"`

	AvgGoalsScored        *float64 `gorm:"column:media_gols_marcados"`
	AvgGoalsConceded      *float64 `gorm:"column:media_gols_sofridos"`
	AvgCornersScored      *float64 `gorm:"column:media_cantos_marcados"`
	AvgCornersConceded    *float64 `gorm:"column:media_cantos_sofridos"`
	FrequencyBTTS         *float64 `gorm:"column:frequencia_btts"`
	FrequencyOver05HT     *float64 `gorm:"column:frequencia_over05ht"`
	FrequencyOver15HT     *float64 `gorm:"column:frequencia_over15ht"`
	FrequencyOver15       *float64 `gorm:"column:frequencia_over15"`
	FrequencyOver25       *float64 `gorm:"column:frequencia_over25"`
	FrequencyOver35       *float64 `gorm:"column:frequencia_over35"`
	FrequencyCorners85    *float64 `gorm:"column:frequencia_cantos85"`

	AvgGoalsScoredHome        *float64 `gorm:"column:media_gols_marcados_mandante"`
	AvgGoalsConcededHome      *float64 `gorm:"column:media_gols_sofridos_mandante"`
	AvgCornersScoredHome      *float64 `gorm:"column:media_cantos_marcados_mandante"`
	AvgCornersConcededHome    *float64 `gorm:"column:media_cantos_sofridos_mandante"`
	FrequencyBTTSHome         *float64 `gorm:"column:frequencia_btts_mandante"`
	FrequencyOver05HTHome     *float64 `gorm:"column:frequencia_over05ht_mandante"`
	FrequencyOver15HTHome     *float64 `gorm:"column:frequencia_over15ht_mandante"`
	FrequencyOver15Home       *float64 `gorm:"column:frequencia_over15_mandante"`
	FrequencyOver25Home       *float64 `gorm:"column:frequencia_over25_mandante"`
	FrequencyOver35Home       *float64 `gorm:"column:frequencia_over35_mandante"`
	FrequencyCorners85Home    *float64 `gorm:"column:frequencia_cantos85_mandante"`

	AvgGoalsScoredAway        *float64 `gorm:"column:media_gols_marcados_visitante"`
	AvgGoalsConcededAway      *float64 `gorm:"column:media_gols_sofridos_visitante"`
	AvgCornersScoredAway      *float64 `gorm:"column:media_cantos_marcados_visitante"`
	AvgCornersConcededAway    *float64 `gorm:"column:media_cantos_sofridos_visitante"`
	FrequencyBTTSAway         *float64 `gorm:"column:frequencia_btts_visitante"`
	FrequencyOver05HTAway     *float64 `gorm:"column:frequencia_over05ht_visitante"`
	FrequencyOver15HTAway     *float64 `gorm:"column:frequencia_over15ht_visitante"`
	FrequencyOver15Away       *float64 `gorm:"column:frequencia_over15_visitante"`
	FrequencyOver25Away       *float64 `gorm:"column:frequencia_over25_visitante"`
	FrequencyOver35Away       *float64 `gorm:"column:frequencia_over35_visitante"`
	FrequencyCorners85Away    *float64 `gorm:"column:frequencia_cantos85_visitante"`

	FrequencyFirstToScoreHome    *float64 `gorm:"column:frequencia_primeiro_marcar_mandante"`
	FrequencyFirstToScoreAway    *float64 `gorm:"column:frequencia_primeiro_marcar_visitante"`
	FrequencyGoal70Home           *float64 `gorm:"column:frequencia_gol_70_mandante"`
	FrequencyGoal70Away           *float64 `gorm:"column:frequencia_gol_70_visitante"`
	FrequencyGoal45Home           *float64 `gorm:"column:frequencia_gol_45_mandante"`
	FrequencyGoal45Away           *float64 `gorm:"column:frequencia_gol_45_visitante"`
	FrequencyGoal30Home           *float64 `gorm:"column:frequencia_gol_30_mandante"`
	FrequencyGoal30Away           *float64 `gorm:"column:frequencia_gol_30_visitante"`
	FrequencyGoal20Home           *float64 `gorm:"column:frequencia_gol_20_mandante"`
	FrequencyGoal20Away           *float64 `gorm:"column:frequencia_gol_20_visitante"`
}

func (TeamStatsEntity) TableName() string {
	return "time_estatisticas"
}

type FormEntry struct {
	MatchID      uint      `json:"match_id"`
	Result       string    `json:"result"`
	OpponentID   uint      `json:"opponent_id"`
	OpponentName string    `json:"opponent_name"`
	HomeScore    int       `json:"home_score"`
	AwayScore    int       `json:"away_score"`
	Venue        string    `json:"venue"`
	Date         time.Time `json:"date"`
}
