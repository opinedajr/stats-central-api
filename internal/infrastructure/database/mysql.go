package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/opinedajr/stats-central-api/internal/shared/config"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type MySQLDatabase struct {
	cfg   config.DatabaseConfig
	log   logger.Logger
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewMySQLDatabase(cfg config.DatabaseConfig, log logger.Logger) *MySQLDatabase {
	return &MySQLDatabase{
		cfg: cfg,
		log: log,
	}
}

func (m *MySQLDatabase) Connect(ctx context.Context) (*gorm.DB, error) {
	if m.db != nil {
		return m.db, nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.cfg.User,
		m.cfg.Password,
		m.cfg.Host,
		m.cfg.Port,
		m.cfg.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		m.log.Error(ctx, "failed to connect to database",
			"host", m.cfg.Host,
			"port", m.cfg.Port,
			"database", m.cfg.Name,
			"error", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := sqlDB.Ping(); err != nil {
		m.log.Error(ctx, "failed to ping database", "error", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	m.log.Info(ctx, "database connection established",
		"host", m.cfg.Host,
		"database", m.cfg.Name)

	m.db = db
	m.sqlDB = sqlDB

	return m.db, nil
}

func (m *MySQLDatabase) Close() error {
	if m.sqlDB != nil {
		return m.sqlDB.Close()
	}
	return nil
}

func (m *MySQLDatabase) Ping() error {
	if m.sqlDB == nil {
		return fmt.Errorf("database not connected")
	}
	return m.sqlDB.Ping()
}

func (m *MySQLDatabase) Migrate(models ...interface{}) error {
	if m.db == nil {
		return fmt.Errorf("database not connected")
	}
	return m.db.AutoMigrate(models...)
}
