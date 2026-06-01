package analyse

import "time"

type AnalyseInput struct {
	TeamID       uint `form:"teamId" binding:"required,min=1"`
	TournamentID uint `form:"tournamentId" binding:"required,min=1"`
	LastN        int  `form:"last_n" binding:"omitempty,min=1,max=50"`
}

type AnalyseOutput struct {
	TeamID       uint          `json:"team_id"`
	TournamentID uint          `json:"tournament_id"`
	Home         VenueContext  `json:"home"`
	Away         VenueContext  `json:"away"`
	Overall      VenueContext  `json:"overall"`
	CalculatedAt time.Time     `json:"calculated_at"`
}

type VenueContext struct {
	Stats              VenueStats   `json:"stats"`
	RecentForm         []FormEntry  `json:"recent_form"`
	RecentFormSummary  FormSummary  `json:"recent_form_summary"`
}

type VenueStats struct {
	MatchesPlayed int      `json:"matches_played"`
	Wins          int      `json:"wins"`
	Draws         int      `json:"draws"`
	Losses        int      `json:"losses"`
	GoalsFor      int      `json:"goals_for"`
	GoalsAgainst  int      `json:"goals_against"`
	WinRate       float64  `json:"win_rate"`

	AvgGoalsScored     *float64 `json:"avg_goals_scored,omitempty"`
	AvgGoalsConceded   *float64 `json:"avg_goals_conceded,omitempty"`
	AvgCornersScored   *float64 `json:"avg_corners_scored,omitempty"`
	AvgCornersConceded *float64 `json:"avg_corners_conceded,omitempty"`
	FrequencyBTTS      *float64 `json:"frequency_btts,omitempty"`
	FrequencyOver05HT  *float64 `json:"frequency_over05ht,omitempty"`
	FrequencyOver15HT  *float64 `json:"frequency_over15ht,omitempty"`
	FrequencyOver15    *float64 `json:"frequency_over15,omitempty"`
	FrequencyOver25    *float64 `json:"frequency_over25,omitempty"`
	FrequencyOver35    *float64 `json:"frequency_over35,omitempty"`
	FrequencyCorners85 *float64 `json:"frequency_corners85,omitempty"`
}

type FormSummary struct {
	MatchesAnalyzed int `json:"matches_analyzed"`
	Wins            int `json:"wins"`
	Draws           int `json:"draws"`
	Losses          int `json:"losses"`
	GoalsFor        int `json:"goals_for"`
	GoalsAgainst    int `json:"goals_against"`
}
