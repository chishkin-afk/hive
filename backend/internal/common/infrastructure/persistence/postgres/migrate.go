package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
)

func Migrate(cfg *config.Config, db *sql.DB) error {
	dirs, err := os.ReadDir(cfg.Persistence.Postgres.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migration's dir: %w", err)
	}

	if err := applyMigrations(cfg, dirs, db); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func applyMigrations(cfg *config.Config, dirs []os.DirEntry, db *sql.DB) error {
	for _, dir := range dirs {
		if dir.IsDir() || !strings.HasSuffix(dir.Name(), ".sql") {
			continue
		}

		bytes, err := os.ReadFile(filepath.Join(cfg.Persistence.Postgres.MigrationsPath, dir.Name()))
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(bytes)); err != nil {
			return err
		}
	}

	return nil
}
