package teams

import (
	"context"

	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

type Service interface {
	ListTeams(ctx context.Context, filter TeamFilter, page int, pageSize int) ([]*TeamOutput, int64, error)
	GetTeamByID(ctx context.Context, id uint) (*TeamOutput, error)
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

func (s *service) ListTeams(ctx context.Context, filter TeamFilter, page int, pageSize int) ([]*TeamOutput, int64, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	teams, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "failed to list teams", "error", err)
		return nil, 0, err
	}

	outputs := make([]*TeamOutput, len(teams))
	for i, t := range teams {
		outputs[i] = toTeamOutput(t)
	}

	return outputs, total, nil
}

func (s *service) GetTeamByID(ctx context.Context, id uint) (*TeamOutput, error) {
	team, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "team not found", "error", err, "team_id", id)
		return nil, err
	}

	return toTeamOutput(team), nil
}

func toTeamOutput(team *Team) *TeamOutput {
	return &TeamOutput{
		ID:          team.ID,
		Name:        team.Name,
		Country:     team.Country,
		SofascoreID: team.SofascoreID,
		SokkerproID: team.SokkerproID,
		CreatedAt:   team.CreatedAt,
		UpdatedAt:   team.UpdatedAt,
	}
}
