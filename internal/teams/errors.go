package teams

import (
	"errors"
	"fmt"
)

var (
	ErrTeamNotFound    = errors.New("team not found")
	ErrDatabaseError   = errors.New("database error")
	ErrValidationFailed = errors.New("validation failed")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
