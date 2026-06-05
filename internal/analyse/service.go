package analyse

import (
	"context"
	"time"

	"github.com/opinedajr/stats-central-api/internal/match"
	"github.com/opinedajr/stats-central-api/internal/teams"
)

const (
	defaultLastN = 10
	maxLastN     = 50
)

type Service interface {
	TeamTournamentAnalysis(ctx context.Context, teamID uint, tournamentID uint, season string, lastN int) (AnalyseOutput, error)
}

type analyseService struct {
	statsRepo   StatsRepository
	matchesRepo match.Repository
	teamsRepo   teams.Repository
}

func NewAnalyseService(statsRepo StatsRepository, matchesRepo match.Repository, teamsRepo teams.Repository) Service {
	return &analyseService{
		statsRepo:   statsRepo,
		matchesRepo: matchesRepo,
		teamsRepo:   teamsRepo,
	}
}

func (s *analyseService) TeamTournamentAnalysis(ctx context.Context, teamID uint, tournamentID uint, season string, lastN int) (AnalyseOutput, error) {
	if lastN <= 0 {
		lastN = defaultLastN
	}
	if lastN > maxLastN {
		lastN = maxLastN
	}

	preCalcStats, err := s.statsRepo.GetTeamStats(ctx, teamID, tournamentID, season)
	if err != nil {
		return AnalyseOutput{}, err
	}

	homeStats, err := s.matchesRepo.GetHomeStats(ctx, teamID, tournamentID, season)
	if err != nil {
		return AnalyseOutput{}, err
	}

	awayStats, err := s.matchesRepo.GetAwayStats(ctx, teamID, tournamentID, season)
	if err != nil {
		return AnalyseOutput{}, err
	}

	overallStats, err := s.matchesRepo.GetOverallStats(ctx, teamID, tournamentID, season)
	if err != nil {
		return AnalyseOutput{}, err
	}

	matches, err := s.matchesRepo.GetRecentMatches(ctx, teamID, tournamentID, season, lastN*3)
	if err != nil {
		return AnalyseOutput{}, err
	}

	homeMatches := filterHome(matches, teamID)
	awayMatches := filterAway(matches, teamID)

	recentOverall := buildRecentForm(ctx, matches, teamID, lastN, s.teamsRepo)
	recentHome := buildRecentForm(ctx, homeMatches, teamID, lastN, s.teamsRepo)
	recentAway := buildRecentForm(ctx, awayMatches, teamID, lastN, s.teamsRepo)

	return AnalyseOutput{
		TeamID:       teamID,
		TournamentID: tournamentID,
		Home: VenueContext{
			Stats:             mapVenueStats(homeStats, preCalcStats, "home"),
			RecentForm:        recentHome,
			RecentFormSummary: summarizeForm(recentHome),
		},
		Away: VenueContext{
			Stats:             mapVenueStats(awayStats, preCalcStats, "away"),
			RecentForm:        recentAway,
			RecentFormSummary: summarizeForm(recentAway),
		},
		Overall: VenueContext{
			Stats:             mapVenueStats(overallStats, preCalcStats, "overall"),
			RecentForm:        recentOverall,
			RecentFormSummary: summarizeForm(recentOverall),
		},
		CalculatedAt: time.Now(),
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

func buildRecentForm(ctx context.Context, matches []*match.MatchEntity, targetTeamID uint, limit int, teamsRepo teams.Repository) []FormEntry {
	if len(matches) > limit {
		matches = matches[:limit]
	}

	entries := make([]FormEntry, 0, len(matches))

	for _, m := range matches {
		var result string
		var homeScore, awayScore int
		var venue string

		if targetTeamID == m.HomeTeamID {
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

		homeTeam, _ := teamsRepo.FindByID(ctx, m.HomeTeamID)
		awayTeam, _ := teamsRepo.FindByID(ctx, m.AwayTeamID)

		homeName := ""
		if homeTeam != nil {
			homeName = homeTeam.Name
		}

		awayName := ""
		if awayTeam != nil {
			awayName = awayTeam.Name
		}

		var date time.Time
		if m.DateTimestamp != nil {
			date = time.Unix(*m.DateTimestamp, 0)
		}

		entries = append(entries, FormEntry{
			MatchID:       m.ID,
			Result:        result,
			HomeID:        m.HomeTeamID,
			HomeName:      homeName,
			AwayID:        m.AwayTeamID,
			AwayName:      awayName,
			HomeScore:     homeScore,
			AwayScore:     awayScore,
			Venue:         venue,
			HomeOdd:       m.HomeTeamOdd,
			AwayOdd:       m.AwayTeamOdd,
			DrawOdd:       m.DrawOdd,
			FirstToScore:  m.FirstToScore,
			SecondToScore: m.SecondToScore,
			ThirdToScore:  m.ThirdToScore,
			Goal1Minute:   m.Goal1Minute,
			Goal2Minute:   m.Goal2Minute,
			Goal3Minute:   m.Goal3Minute,
			Date:          date,
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

	switch venue {
	case "home":
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
		result.FrequencyFirstToScore = preCalcStats.FrequencyFirstToScoreHome
		result.FrequencyGoal20 = preCalcStats.FrequencyGoal20Home
		result.FrequencyGoal45 = preCalcStats.FrequencyGoal45Home
		result.FrequencyGoal70 = preCalcStats.FrequencyGoal70Home
	case "away":
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
		result.FrequencyFirstToScore = preCalcStats.FrequencyFirstToScoreAway
		result.FrequencyGoal20 = preCalcStats.FrequencyGoal20Away
		result.FrequencyGoal45 = preCalcStats.FrequencyGoal45Away
		result.FrequencyGoal70 = preCalcStats.FrequencyGoal70Away
	default:
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

		switch entry.Result {
		case "W":
			summary.Wins++
		case "D":
			summary.Draws++
		default:
			summary.Losses++
		}
	}

	return summary
}
