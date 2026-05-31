package postgresconnect

import (
	"database/sql"
	"fmt"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	_ "github.com/lib/pq"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", getDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with postgres: %w", err)
	}

	applySettings(db, cfg)
	return db, nil
}

func applySettings(db *sql.DB, cfg *config.Config) {
	db.SetConnMaxIdleTime(cfg.Persistence.Postgres.Conns.MaxIdleTime)
	db.SetConnMaxLifetime(cfg.Persistence.Postgres.Conns.MaxLifetime)
	db.SetMaxIdleConns(cfg.Persistence.Postgres.Conns.MaxIdles)
	db.SetMaxOpenConns(cfg.Persistence.Postgres.Conns.MaxOpens)
}

func getDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%d sslmode=%s user=%s password=%s dbname=%s",
		cfg.Persistence.Postgres.Host,
		cfg.Persistence.Postgres.Port,
		cfg.Persistence.Postgres.SSLMode,
		cfg.Persistence.Postgres.Auth.User,
		cfg.Persistence.Postgres.Auth.Password,
		cfg.Persistence.Postgres.Auth.DBName,
	)
}
