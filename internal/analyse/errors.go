package analyse

import (
	"errors"
	"fmt"
)

var (
	ErrTeamNotFound     = errors.New("team not found")
	ErrTournamentNotFound = errors.New("tournament not found")
	ErrStatsNotFound    = errors.New("stats not found")
	ErrInvalidLastN     = errors.New("invalid last_n")
	ErrDatabaseError    = errors.New("database error")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
