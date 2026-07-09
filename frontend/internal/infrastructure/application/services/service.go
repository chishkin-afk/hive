package services

import (
	"fmt"
	"log/slog"

	"github.com/devathh/hive/frontend/internal/infrastructure/config"
)

type FrontendServices struct {
	cfg *config.Config
	log *slog.Logger
}

func New(
	cfg *config.Config,
	log *slog.Logger,
) *FrontendServices {
	return &FrontendServices{
		cfg: cfg,
		log: log,
	}
}

func (fs *FrontendServices) Config() []byte {
	str := fs.stringJSConfig()
	return []byte(str)
}

func (fs *FrontendServices) stringJSConfig() string {
	return fmt.Sprintf(`
window.APP_CONFIG = {
    apiUrl: "%s",
    version: "%s"
};
	`, fs.cfg.Backend.BaseURL, fs.cfg.App.Version)
}
