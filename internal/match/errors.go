package match

import (
	"errors"
	"fmt"
)

var (
	ErrDatabaseError = errors.New("database error")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
