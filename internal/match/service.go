package match

import (
	"context"

	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/opinedajr/stats-central-api/internal/shared/pagination"
)

type Service interface {
	ListMatches(ctx context.Context, filter MatchFilter, page int, pageSize int) ([]*MatchOutput, int64, error)
}

type service struct {
	repo   Repository
	logger logger.Logger
}

func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) ListMatches(ctx context.Context, filter MatchFilter, page int, pageSize int) ([]*MatchOutput, int64, error) {
	page, pageSize = pagination.Normalize(page, pageSize)

	matches, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "failed to list matches", "error", err)
		return nil, 0, err
	}

	outputs := make([]*MatchOutput, len(matches))
	for i, m := range matches {
		outputs[i] = toMatchOutput(m)
	}

	return outputs, total, nil
}

func toMatchOutput(m *MatchEntity) *MatchOutput {
	return &MatchOutput{
		ID:             m.ID,
		TournamentID:   m.LeagueID,
		Season:         m.Season,
		Round:          m.Round,
		DateTimestamp:  m.DateTimestamp,
		Status:         m.Status,
		Time:           m.Time,
		HomeTeamID:     m.HomeTeamID,
		HomeTeamName:   m.HomeTeamName,
		HomeTeamGoals:  m.HomeTeamGoals,
		HomeTeamOdd:    m.HomeTeamOdd,
		AwayTeamID:     m.AwayTeamID,
		AwayTeamName:   m.AwayTeamName,
		AwayTeamGoals:  m.AwayTeamGoals,
		AwayTeamOdd:    m.AwayTeamOdd,
		DrawOdd:        m.DrawOdd,
		BTTSOdd:        m.BTTSOdd,
		Under25Odd:     m.Under25Odd,
	}
}
