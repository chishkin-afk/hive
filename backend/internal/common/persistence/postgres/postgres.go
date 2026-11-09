package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/chishkin-afk/hive/backend/internal/common/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg *config.Persistence) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(getDataSourceName("postgres", cfg))
	if err != nil {
		return nil, fmt.Errorf("can't get pgx pool config: %w", err)
	}

	config.MaxConns = int32(cfg.Postgres.Conns.Max)
	config.MinConns = int32(cfg.Postgres.Conns.Min)
	config.MaxConnIdleTime = cfg.Postgres.Conns.MaxIdleTime
	config.MaxConnLifetime = cfg.Postgres.Conns.MaxLifetime

	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("can't create pgx conn: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("can't ping pgx conn: %w", err)
	}

	return conn, nil
}

func Migrate(cfg *config.Persistence) error {
	migrator, err := migrate.New("file://"+cfg.Migrations, getDataSourceName("pgx5", cfg))
	if err != nil {
		return fmt.Errorf("can't create migrate: %w", err)
	}
	defer func() {
		_, _ = migrator.Close()
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("can't up migrations: %w", err)
	}

	return nil
}

func getDataSourceName(schema string, cfg *config.Persistence) string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		schema,
		cfg.Postgres.Auth.User,
		cfg.Postgres.Auth.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Auth.DB,
		cfg.Postgres.SSLMode,
	)
}
