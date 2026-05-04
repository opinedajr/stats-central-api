package tournament

import "context"

type TournamentRepository interface {
	Create(ctx context.Context, tournament *Tournament) error
	Update(ctx context.Context, tournament *Tournament) error
	UpdateStatus(ctx context.Context, id uint, active bool) error
	FindByID(ctx context.Context, id uint) (*Tournament, error)
	List(ctx context.Context, filter TournamentFilter, page int, pageSize int) ([]*Tournament, int64, error)
}
