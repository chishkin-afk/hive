package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type Config struct {
	App         App         `yaml:"app"`
	Middlewares Middlewares `yaml:"middlewares"`
	Server      Server      `yaml:"server"`
	Backend     Backend     `yaml:"backend"`
}

type App struct {
	Name    string `yaml:"name" validate:"required"`
	Version string `yaml:"version" validate:"required,semver"`
	Env     string `yaml:"env" validate:"required,oneof=dev prod local"`
}

type Middlewares struct {
	RateLimit struct {
		Window time.Duration `yaml:"window" validate:"required,min=100ms"`
		Limit  int           `yaml:"limit" validate:"required,gte=1"`
	} `yaml:"rate_limit"`
}

type Server struct {
	HTTP struct {
		Addr string `yaml:"addr" validate:"required,hostname_port"`
		TLS  struct {
			Enable         bool   `yaml:"enable"`
			ServerCertPath string `yaml:"server_cert_path"`
			ServerKeyPath  string `yaml:"server_key_path"`
		} `yaml:"tls"`
	} `yaml:"http"`
	Conns struct {
		ReadTimeout  time.Duration `yaml:"read_timeout" validate:"required,min=100ms"`
		WriteTimeout time.Duration `yaml:"write_timeout" validate:"required,min=100ms"`
		IdleTimeout  time.Duration `yaml:"idle_timeout" validate:"required,min=100ms"`
	} `yaml:"conns"`
}

type Backend struct {
	BaseURL string `yaml:"base_url" validate:"required,url"`
}

func New(path string) (*Config, error) {
	bytes, err := readFile(path)
	if err != nil {
		return nil, err
	}

	cfg, err := parseBytes(bytes)
	if err != nil {
		return nil, err
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return err
	}

	return validateTLS(cfg)
}

func validateTLS(cfg *Config) error {
	if !cfg.Server.HTTP.TLS.Enable {
		return nil
	}

	path := filepath.Clean(cfg.Server.HTTP.TLS.ServerCertPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("invalid server cert path: %w", err)
	}

	path = filepath.Clean(cfg.Server.HTTP.TLS.ServerKeyPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("invalid server key path: %w", err)
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

func validatePath(path string) error {
	if _, err := filepath.Abs(path); err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	return nil
}

func readFile(path string) ([]byte, error) {
	path = filepath.Clean(path)
	if err := validatePath(path); err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := os.ExpandEnv(string(bytes))
	return []byte(content), nil
}
