package teams

import "context"

type Repository interface {
	FindByID(ctx context.Context, id uint) (*Team, error)
	List(ctx context.Context, filter TeamFilter, page int, pageSize int) ([]*Team, int64, error)
}
