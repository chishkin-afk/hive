package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/a8m/envsubst"
	"github.com/go-playground/validator/v10"
	"go.yaml.in/yaml/v3"
)

var (
	ErrInvalidCertPath = errors.New("invalid path to server cert")
	ErrInvalidKeyPath  = errors.New("invalid path to server key")
)

type Config struct {
	App         App         `yaml:"app"`
	Server      Server      `yaml:"server"`
	Persistence Persistence `yaml:"persistence"`
	Cache       Cache       `yaml:"cache"`
}

type App struct {
	Name    string `yaml:"name" validate:"required"`
	Version string `yaml:"version" validate:"required,semver"`
	Env     string `yaml:"env" validate:"required,oneof=dev local prod"`
}

type Server struct {
	HTTP struct {
		Host string `yaml:"host" validate:"required,hostname"`
		Port int    `yaml:"port" validate:"required,port"`
		// TLS fields can be empty, its depends on enabled field
		//
		// The validation of this fields will be in manual function
		TLS struct {
			Enabled  bool   `yaml:"enabled"`
			CertPath string `yaml:"cert_path" validate:"omitempty"`
			KeyPath  string `yaml:"key_path" validate:"omitempty"`
		} `yaml:"tls"`
	} `yaml:"http"`
	Timeouts struct {
		Read  time.Duration `yaml:"read" validate:"required,min=100ms"`
		Write time.Duration `yaml:"write" validate:"required,min=100ms"`
		Idle  time.Duration `yaml:"idle" validate:"required,min=100ms"`
	} `yaml:"timeouts"`
}

type Persistence struct {
	Migrations string `yaml:"migrations" validate:"required"`
	Postgres   struct {
		Host    string `yaml:"host" validate:"required,hostname"`
		Port    int    `yaml:"port" validate:"required,port"`
		SSLMode string `yaml:"sslmode" validate:"required,oneof=disable verify-full allow require prefer verify-ca"`
		Auth    struct {
			User     string `yaml:"user" validate:"required"`
			Password string `yaml:"password" validate:"required"`
			DB       string `yaml:"db" validate:"required"`
		}
		Conns struct {
			Max         int           `yaml:"max" validate:"required,gte=1"`
			Min         int           `yaml:"min" validate:"required,gte=1"`
			MaxIdleTime time.Duration `yaml:"max_idle_time" validate:"required,min=1m"`
			MaxLifetime time.Duration `yaml:"max_lifetime" validate:"required,min=1m"`
		} `yaml:"conns"`
	} `yaml:"postgres"`
}

type Cache struct {
	Redis struct {
		Host string `yaml:"host" validate:"required,hostname"`
		Port int    `yaml:"port" validate:"required,port"`
		// Auth fields may be empty, it means that defaults settings will be used
		Auth struct {
			Username string `yaml:"username" validate:"omitempty"`
			Password string `yaml:"password" validate:"omitempty"`
			DB       int    `yaml:"db" validate:"gte=0,lte=15"`
		} `yaml:"auth"`
	} `yaml:"redis"`
}

// New is a function, for load and parse config file
//
// Calling new means that .env file was loaded
func New(path string) (*Config, error) {
	bytes, err := loadBytes(path)
	if err != nil {
		return nil, fmt.Errorf("can't load config file bytes: %w", err)
	}

	cfg, err := parseBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("can't parse config file bytes: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return err
	}

	return validateTLS(&cfg.Server)
}

func validateTLS(cfg *Server) error {
	if cfg == nil || !cfg.HTTP.TLS.Enabled {
		return nil
	}

	certPath := filepath.Clean(cfg.HTTP.TLS.CertPath)
	if stat, err := os.Stat(certPath); err != nil && stat.IsDir() {
		return ErrInvalidCertPath
	}

	keyPath := filepath.Clean(cfg.HTTP.TLS.KeyPath)
	if stat, err := os.Stat(keyPath); err != nil && stat.IsDir() {
		return ErrInvalidKeyPath
	}

	return nil
}

func parseBytes(bytes []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadBytes(path string) ([]byte, error) {
	path = filepath.Clean(path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content, err := envsubst.Bytes(bytes)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}
