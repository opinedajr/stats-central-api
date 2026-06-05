package tournament

import (
	"context"

	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"github.com/opinedajr/stats-central-api/internal/shared/pagination"
)

type Service interface {
	CreateTournament(ctx context.Context, input CreateTournamentInput) (*TournamentOutput, error)
	ListTournaments(ctx context.Context, filter TournamentFilter, page int, pageSize int) ([]*TournamentOutput, int64, error)
	GetTournamentByID(ctx context.Context, id uint) (*TournamentOutput, error)
	UpdateTournament(ctx context.Context, id uint, input UpdateTournamentInput) (*TournamentOutput, error)
	UpdateTournamentStatus(ctx context.Context, id uint, active bool) (*TournamentOutput, error)
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

func (s *service) CreateTournament(ctx context.Context, input CreateTournamentInput) (*TournamentOutput, error) {
	active := true
	if input.Active != nil {
		active = *input.Active
	}

	tournament := &Tournament{
		Name:     input.Name,
		Country:  input.Country,
		Division: input.Division,
		Season:   input.Season,
		Round:    input.Round,
		Active:   active,
	}

	if err := s.repo.Create(ctx, tournament); err != nil {
		s.logger.Error(ctx, "failed to create tournament", "error", err, "name", input.Name)
		return nil, err
	}

	s.logger.Info(ctx, "tournament created", "tournament_id", tournament.ID, "name", input.Name, "country", input.Country, "active", active, "outcome", "success")

	return toTournamentOutput(tournament), nil
}

func (s *service) ListTournaments(ctx context.Context, filter TournamentFilter, page int, pageSize int) ([]*TournamentOutput, int64, error) {
	page, pageSize = pagination.Normalize(page, pageSize)

	tournaments, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "failed to list tournaments", "error", err)
		return nil, 0, err
	}

	outputs := make([]*TournamentOutput, len(tournaments))
	for i, t := range tournaments {
		outputs[i] = toTournamentOutput(t)
	}

	return outputs, total, nil
}

func (s *service) GetTournamentByID(ctx context.Context, id uint) (*TournamentOutput, error) {
	tournament, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "tournament not found", "error", err, "tournament_id", id)
		return nil, err
	}

	return toTournamentOutput(tournament), nil
}

func (s *service) UpdateTournament(ctx context.Context, id uint, input UpdateTournamentInput) (*TournamentOutput, error) {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "tournament not found", "error", err, "tournament_id", id)
		return nil, err
	}

	tournament := &Tournament{
		ID:       id,
		Name:     input.Name,
		Country:  input.Country,
		Division: input.Division,
		Season:   input.Season,
		Round:    input.Round,
	}

	if err := s.repo.Update(ctx, tournament); err != nil {
		s.logger.Error(ctx, "failed to update tournament", "error", err, "tournament_id", id)
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve updated tournament", "error", err, "tournament_id", id)
		return nil, err
	}

	s.logger.Info(ctx, "tournament updated", "tournament_id", id, "name", input.Name, "country", input.Country, "outcome", "success")

	return toTournamentOutput(updated), nil
}

func (s *service) UpdateTournamentStatus(ctx context.Context, id uint, active bool) (*TournamentOutput, error) {
	if err := s.repo.UpdateStatus(ctx, id, active); err != nil {
		s.logger.Error(ctx, "failed to update tournament status", "error", err, "tournament_id", id, "active", active)
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to retrieve updated tournament", "error", err, "tournament_id", id)
		return nil, err
	}

	s.logger.Info(ctx, "tournament status updated", "tournament_id", id, "active", active, "outcome", "success")

	return toTournamentOutput(updated), nil
}

func toTournamentOutput(tournament *Tournament) *TournamentOutput {
	return &TournamentOutput{
		ID:        tournament.ID,
		Name:      tournament.Name,
		Country:   tournament.Country,
		Division:  tournament.Division,
		Season:    tournament.Season,
		Round:     tournament.Round,
		Active:    tournament.Active,
		CreatedAt: tournament.CreatedAt,
		UpdatedAt: tournament.UpdatedAt,
	}
}
