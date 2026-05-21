package analyse

import (
	"context"
	"time"

	"github.com/opinedajr/stats-central-api/internal/match"
)

type Service interface {
	TeamTournamentAnalysis(ctx context.Context, teamID uint, tournamentID uint, lastN int) (AnalyseOutput, error)
}

type analyseService struct {
	statsRepo   StatsRepository
	matchesRepo match.MatchRepository
}

func NewAnalyseService(statsRepo StatsRepository, matchesRepo match.MatchRepository) Service {
	return &analyseService{
		statsRepo:   statsRepo,
		matchesRepo: matchesRepo,
	}
}

func (s *analyseService) TeamTournamentAnalysis(ctx context.Context, teamID uint, tournamentID uint, lastN int) (AnalyseOutput, error) {
	if lastN <= 0 {
		lastN = 10
	}
	if lastN > 50 {
		lastN = 50
	}

	preCalcStats, err := s.statsRepo.GetTeamStats(ctx, teamID, tournamentID)
	if err != nil {
		return AnalyseOutput{}, err
	}

	homeStats, err := s.matchesRepo.GetHomeStats(ctx, teamID, tournamentID)
	if err != nil {
		return AnalyseOutput{}, err
	}

	awayStats, err := s.matchesRepo.GetAwayStats(ctx, teamID, tournamentID)
	if err != nil {
		return AnalyseOutput{}, err
	}

	overallStats, err := s.matchesRepo.GetOverallStats(ctx, teamID, tournamentID)
	if err != nil {
		return AnalyseOutput{}, err
	}

	matches, err := s.matchesRepo.GetRecentMatches(ctx, teamID, tournamentID, lastN*3)
	if err != nil {
		return AnalyseOutput{}, err
	}

	homeMatches := filterHome(matches, teamID)
	awayMatches := filterAway(matches, teamID)

	recentOverall := buildRecentForm(matches, teamID, lastN)
	recentHome := buildRecentForm(homeMatches, teamID, lastN)
	recentAway := buildRecentForm(awayMatches, teamID, lastN)

	return AnalyseOutput{
		TeamID:       teamID,
		TournamentID: tournamentID,
		HomeStats:    mapVenueStats(homeStats, preCalcStats, "home"),
		AwayStats:    mapVenueStats(awayStats, preCalcStats, "away"),
		OverallStats: mapVenueStats(overallStats, preCalcStats, "overall"),
		RecentForm:           recentOverall,
		RecentFormSummary:    summarizeForm(recentOverall),
		RecentFormHome:       recentHome,
		RecentFormHomeSummary: summarizeForm(recentHome),
		RecentFormAway:       recentAway,
		RecentFormAwaySummary: summarizeForm(recentAway),
		CalculatedAt:  time.Now(),
	}, nil
}

func filterHome(matches []*match.MatchEntity, teamID uint) []*match.MatchEntity {
	home := make([]*match.MatchEntity, 0)
	for _, m := range matches {
		if m.HomeTeamID == teamID {
			home = append(home, m)
		}
	}
	return home
}

func filterAway(matches []*match.MatchEntity, teamID uint) []*match.MatchEntity {
	away := make([]*match.MatchEntity, 0)
	for _, m := range matches {
		if m.AwayTeamID == teamID {
			away = append(away, m)
		}
	}
	return away
}

func buildRecentForm(matches []*match.MatchEntity, targetTeamID uint, limit int) []FormEntry {
	if len(matches) > limit {
		matches = matches[:limit]
	}

	entries := make([]FormEntry, 0, len(matches))

	for _, m := range matches {
		var result string
		var opponentID uint
		var homeScore, awayScore int
		var venue string

		if targetTeamID == m.HomeTeamID {
			opponentID = m.AwayTeamID
			homeScore = m.HomeTeamGoals
			awayScore = m.AwayTeamGoals
			venue = "home"

			if homeScore > awayScore {
				result = "W"
			} else if homeScore == awayScore {
				result = "D"
			} else {
				result = "L"
			}
		} else {
			opponentID = m.HomeTeamID
			homeScore = m.HomeTeamGoals
			awayScore = m.AwayTeamGoals
			venue = "away"

			if awayScore > homeScore {
				result = "W"
			} else if awayScore == homeScore {
				result = "D"
			} else {
				result = "L"
			}
		}

		date := time.Unix(*m.DateTimestamp, 0)

		entries = append(entries, FormEntry{
			MatchID:      m.ID,
			Result:       result,
			OpponentID:   opponentID,
			OpponentName: "",
			HomeScore:    homeScore,
			AwayScore:    awayScore,
			Venue:        venue,
			Date:         date,
		})
	}

	return entries
}

func mapVenueStats(calcStats match.VenueStatsEntity, preCalcStats *TeamStatsEntity, venue string) VenueStats {
	winRate := 0.0
	if calcStats.MatchesPlayed > 0 {
		winRate = float64(calcStats.Wins) / float64(calcStats.MatchesPlayed)
	}

	result := VenueStats{
		MatchesPlayed: calcStats.MatchesPlayed,
		Wins:          calcStats.Wins,
		Draws:         calcStats.Draws,
		Losses:        calcStats.Losses,
		GoalsFor:      calcStats.GoalsFor,
		GoalsAgainst:  calcStats.GoalsAgainst,
		WinRate:       winRate,
	}

	if venue == "home" {
		result.AvgGoalsScored = preCalcStats.AvgGoalsScoredHome
		result.AvgGoalsConceded = preCalcStats.AvgGoalsConcededHome
		result.AvgCornersScored = preCalcStats.AvgCornersScoredHome
		result.AvgCornersConceded = preCalcStats.AvgCornersConcededHome
		result.FrequencyBTTS = preCalcStats.FrequencyBTTSHome
		result.FrequencyOver05HT = preCalcStats.FrequencyOver05HTHome
		result.FrequencyOver15HT = preCalcStats.FrequencyOver15HTHome
		result.FrequencyOver15 = preCalcStats.FrequencyOver15Home
		result.FrequencyOver25 = preCalcStats.FrequencyOver25Home
		result.FrequencyOver35 = preCalcStats.FrequencyOver35Home
		result.FrequencyCorners85 = preCalcStats.FrequencyCorners85Home
	} else if venue == "away" {
		result.AvgGoalsScored = preCalcStats.AvgGoalsScoredAway
		result.AvgGoalsConceded = preCalcStats.AvgGoalsConcededAway
		result.AvgCornersScored = preCalcStats.AvgCornersScoredAway
		result.AvgCornersConceded = preCalcStats.AvgCornersConcededAway
		result.FrequencyBTTS = preCalcStats.FrequencyBTTSAway
		result.FrequencyOver05HT = preCalcStats.FrequencyOver05HTAway
		result.FrequencyOver15HT = preCalcStats.FrequencyOver15HTAway
		result.FrequencyOver15 = preCalcStats.FrequencyOver15Away
		result.FrequencyOver25 = preCalcStats.FrequencyOver25Away
		result.FrequencyOver35 = preCalcStats.FrequencyOver35Away
		result.FrequencyCorners85 = preCalcStats.FrequencyCorners85Away
	} else {
		result.AvgGoalsScored = preCalcStats.AvgGoalsScored
		result.AvgGoalsConceded = preCalcStats.AvgGoalsConceded
		result.AvgCornersScored = preCalcStats.AvgCornersScored
		result.AvgCornersConceded = preCalcStats.AvgCornersConceded
		result.FrequencyBTTS = preCalcStats.FrequencyBTTS
		result.FrequencyOver05HT = preCalcStats.FrequencyOver05HT
		result.FrequencyOver15HT = preCalcStats.FrequencyOver15HT
		result.FrequencyOver15 = preCalcStats.FrequencyOver15
		result.FrequencyOver25 = preCalcStats.FrequencyOver25
		result.FrequencyOver35 = preCalcStats.FrequencyOver35
		result.FrequencyCorners85 = preCalcStats.FrequencyCorners85
	}

	return result
}

func summarizeForm(form []FormEntry) FormSummary {
	summary := FormSummary{}

	for _, entry := range form {
		summary.MatchesAnalyzed++
		summary.GoalsFor += entry.HomeScore
		summary.GoalsAgainst += entry.AwayScore

		if entry.Result == "W" {
			summary.Wins++
		} else if entry.Result == "D" {
			summary.Draws++
		} else {
			summary.Losses++
		}
	}

	return summary
}
