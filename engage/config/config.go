package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

const (
	Debug Env = "debug"
	Prod  Env = "prod"

	OutputTypeConsole OutputType = "console"
	OutputTypeFile    OutputType = "file"
	OutputTypeBoth    OutputType = "both"
)

type Env string
type OutputType string

type HTTPAuthConfig struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type HTTPConfig struct {
	Host         string         `yaml:"host"`
	Port         int            `yaml:"port"`
	ReadTimeout  time.Duration  `yaml:"read_timeout"`
	WriteTimeout time.Duration  `yaml:"write_timeout"`
	IdleTimeout  time.Duration  `yaml:"idle_timeout"`
	Auth         HTTPAuthConfig `yaml:"auth"`
}

type TgWebhookConfig struct {
	URL    string `yaml:"url"`
	Secret string `yaml:"secret"`
}

type TgConfig struct {
	Token           string          `yaml:"token"`
	SupergroudID    int64           `yaml:"supergroup-id"`
	TicketsThreadID int             `yaml:"tickets-thread-id"`
	Webhook         TgWebhookConfig `yaml:"webhook"`
}

type CacheConfig struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Password    string        `yaml:"password"`
	DB          int           `yaml:"db"`
	PoolSize    int           `yaml:"pool-size"`
	DialTimeout time.Duration `yaml:"dial-timeout"`
	MaxRetries  int           `yaml:"max-retries"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	MaxConns int32  `yaml:"max-conns"`
	MinConns int32  `yaml:"min-conns"`
}

type LogConfig struct {
	Service     string     `yaml:"service"`
	OutputType  OutputType `yaml:"output-type"`
	LogFilePath string     `yaml:"log-file-path"`
	JSON        bool       `yaml:"json"`
}

type Config struct {
	Env             Env           `yaml:"env"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	HTTP            HTTPConfig    `yaml:"http"`
	Tg              TgConfig      `yaml:"tg"`
	Cache           CacheConfig   `yaml:"cache"`
	DB              DBConfig      `yaml:"db"`
	Log             LogConfig     `yaml:"log"`
}

func (c *Config) validate() error {
	var errs []error

	if c.Env == "" {
		c.Env = Debug
	}
	if c.Env != Debug && c.Env != Prod {
		errs = append(errs, fmt.Errorf("env must be 'debug' or 'prod', got '%s'", c.Env))
	}

	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 5 * time.Second
	}

	return errors.Join(errs...)
}

func (a *HTTPAuthConfig) validate() error {
	var errs []error

	if a.Host == "" {
		errs = append(errs, errors.New("http.auth.host is required"))
	}
	if a.Port <= 0 {
		errs = append(errs, errors.New("http.auth.port is required"))
	}
	if a.Timeout == 0 {
		a.Timeout = 5 * time.Second
	}

	return errors.Join(errs...)
}

func (h *HTTPConfig) validate() error {
	var errs []error

	if h.Host == "" {
		errs = append(errs, errors.New("http.host is required"))
	}
	if h.Port <= 0 || h.Port > 65535 {
		errs = append(errs, errors.New("http.port must be between 1 and 65535"))
	}
	if h.ReadTimeout == 0 {
		h.ReadTimeout = 15 * time.Second
	}
	if h.WriteTimeout == 0 {
		h.WriteTimeout = 15 * time.Second
	}
	if h.IdleTimeout == 0 {
		h.IdleTimeout = 60 * time.Second
	}

	if err := h.Auth.validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (t *TgConfig) validate(env Env) error {
	var errs []error

	if t.Token == "" {
		errs = append(errs, errors.New("tg.token is required"))
	}
	if t.SupergroudID == 0 {
		errs = append(errs, errors.New("tg.supergroup_id is required"))
	}
	if t.TicketsThreadID == 0 {
		errs = append(errs, errors.New("tg.tickets_thread_id is required"))
	}

	if env == Prod {
		if t.Webhook.URL == "" {
			errs = append(errs, errors.New("tg.webhook.url is required in prod env"))
		} else if _, err := url.ParseRequestURI(t.Webhook.URL); err != nil {
			errs = append(errs, fmt.Errorf("tg.webhook.url is invalid: %w", err))
		}
		if t.Webhook.Secret == "" {
			errs = append(errs, errors.New("tg.webhook.secret is required in prod env"))
		}
	}

	return errors.Join(errs...)
}

func (c *CacheConfig) validate() error {
	var errs []error

	if c.Host == "" {
		errs = append(errs, errors.New("cache.host is required"))
	}
	if c.Port <= 0 || c.Port > 65535 {
		errs = append(errs, errors.New("cache.port must be between 1 and 65535"))
	}
	if c.PoolSize == 0 {
		c.PoolSize = 10
	}
	if c.DialTimeout == 0 {
		c.DialTimeout = 5 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}

	return errors.Join(errs...)
}

func (d *DBConfig) validate() error {
	var errs []error

	if d.Host == "" {
		errs = append(errs, errors.New("db.host is required"))
	}
	if d.Port <= 0 || d.Port > 65535 {
		errs = append(errs, errors.New("db.port must be between 1 and 65535"))
	}
	if d.DBName == "" {
		errs = append(errs, errors.New("db.dbname is required"))
	}
	if d.MaxConns == 0 {
		d.MaxConns = 20
	}
	if d.MinConns == 0 {
		d.MinConns = 2
	}

	return errors.Join(errs...)
}

func (l *LogConfig) validate() error {
	if l.OutputType == "" {
		l.OutputType = OutputTypeConsole
	}
	if l.OutputType != OutputTypeConsole &&
		l.OutputType != OutputTypeFile &&
		l.OutputType != OutputTypeBoth {
		return fmt.Errorf("log.output_type must be 'console', 'file' or 'both', got '%s'", l.OutputType)
	}

	return nil
}

func (cfg *Config) Validate() error {
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("root config error: %w", err)
	}
	if err := cfg.HTTP.validate(); err != nil {
		return fmt.Errorf("http config error: %w", err)
	}
	if err := cfg.Tg.validate(cfg.Env); err != nil {
		return fmt.Errorf("tg config error: %w", err)
	}
	if err := cfg.Cache.validate(); err != nil {
		return fmt.Errorf("cache config error: %w", err)
	}
	if err := cfg.DB.validate(); err != nil {
		return fmt.Errorf("db config error: %w", err)
	}
	if err := cfg.Log.validate(); err != nil {
		return fmt.Errorf("log config error: %w", err)
	}

	return nil
}

func New(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	expandedData := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expandedData), &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (db *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.DBName,
	)
}

func (db *DBConfig) URL() string {
	return fmt.Sprintf(
		"%s:%s@%s:%d/%s?sslmode=disable",
		db.User, db.Password, db.Host, db.Port, db.DBName,
	)
}
