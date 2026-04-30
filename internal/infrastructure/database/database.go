package database

import (
	"context"

	"gorm.io/gorm"
)

type DatabaseConnection interface {
	Connect(ctx context.Context) (*gorm.DB, error)
	Close() error
	Ping() error
}

type TestDatabaseConnection interface {
	DatabaseConnection
	Migrate(models ...interface{}) error
}
