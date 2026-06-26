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
	App           App           `yaml:"app"`
	Service       Service       `yaml:"service"`
	Middlewares   Middlewares   `yaml:"middlewares"`
	Server        Server        `yaml:"server"`
	Persistence   Persistence   `yaml:"persistence"`
	Cache         Cache         `yaml:"cache"`
	Authorization Authorization `yaml:"authorization"`
	SMTP          SMTP          `yaml:"smtp"`
	Yookassa      Yookassa      `yaml:"yookassa"`
}

type App struct {
	Name    string `yaml:"name" validate:"required"`
	Version string `yaml:"version" validate:"required,semver"`
	Env     string `yaml:"env" validate:"required,oneof=dev staging prod"`
}

type Service struct {
	TrialPeriod      int `yaml:"trial_period_days" validate:"required,gte=1"`
	MaxRegistrations int `yaml:"max_registrations" validate:"required,gte=1"`
}

type Server struct {
	HTTP struct {
		Addr string `yaml:"addr" validate:"required,hostname_port"`
		// TLS will be checked separate from others
		// if enable is true
		TLS struct {
			Enable         bool   `yaml:"enable"`
			ServerCertPath string `yaml:"server_cert_path" validate:"omitempty"`
			ServerKeyPath  string `yaml:"server_key_path" validate:"omitempty"`
		} `yaml:"tls"`
	} `yaml:"http"`
	Conns struct {
		ReadTimeout  time.Duration `yaml:"read_timeout" validate:"required,min=100ms"`
		WriteTimeout time.Duration `yaml:"write_timeout" validate:"required,min=100ms"`
		IdleTimeout  time.Duration `yaml:"idle_timeout" validate:"required,min=100ms"`
	} `yaml:"conns"`
}

type Persistence struct {
	Postgres struct {
		MigrationsPath string `yaml:"migrations_path" validate:"required,dir"`
		Host           string `yaml:"host" validate:"required,hostname"`
		Port           int    `yaml:"port" validate:"required,gte=1,lte=65535"`
		SSLMode        string `yaml:"sslmode" validate:"required,oneof=disable enable"`
		Auth           struct {
			User     string `yaml:"user" validate:"required"`
			Password string `yaml:"password" validate:"required"`
			DBName   string `yaml:"dbname" validate:"required"`
		} `yaml:"auth"`
		Conns struct {
			MaxIdles    int           `yaml:"max_idles" validate:"required,gte=2"`
			MaxIdleTime time.Duration `yaml:"max_idle_time" validate:"required,min=100ms"`
			MaxOpens    int           `yaml:"max_opens" validate:"required,gte=1"`
			MaxLifetime time.Duration `yaml:"max_lifetime" validate:"required,min=100ms"`
		} `yaml:"conns"`
	} `yaml:"postgres"`
}

type Cache struct {
	UserTTL      time.Duration `yaml:"user_ttl" validate:"required,min=1m"`
	OfferTTL     time.Duration `yaml:"offer_ttl" validate:"required,min=1m"`
	ClientTTL    time.Duration `yaml:"client_ttl" validate:"required,min=1m"`
	SchedulesTTL time.Duration `yaml:"schedules_ttl" validate:"required,min=1m"`
	BookingTTL   time.Duration `yaml:"booking_ttl" validate:"required,min=1m"`
	PaymentTTL   time.Duration `yaml:"payment_ttl" validate:"required,min=1m"`
	RateTTL      time.Duration `yaml:"rate_ttl" validate:"required,min=1m"`
	PreparingTTL time.Duration `yaml:"preparing_ttl" validate:"required,min=1m"`
	Redis        struct {
		Host string `yaml:"host" validate:"required,hostname"`
		Port int    `yaml:"port" validate:"required,gte=1,lte=65535"`
		Auth struct {
			Username string `yaml:"username" validate:"omitempty"`
			Password string `yaml:"password" validate:"omitempty"`
			DB       int    `yaml:"db" validate:"omitempty"`
		} `yaml:"auth"`
	} `yaml:"redis"`
}

type Authorization struct {
	JWT struct {
		TokenTTL       time.Duration `yaml:"token_ttl" validate:"required,min=1h"`
		PublicKeyPath  string        `yaml:"public_key_path" validate:"required,file"`
		PrivateKeyPath string        `yaml:"private_key_path" validate:"required,file"`
	} `yaml:"jwt"`
}

type Yookassa struct {
	BaseURL string `yaml:"base_url" validate:"required,url"`
	Auth    struct {
		ShopID    string `yaml:"shop_id" validate:"required"`
		SecretKey string `yaml:"secret_key" validate:"required"`
	} `yaml:"auth"`
	Creates struct {
		Capture      bool   `yaml:"capture"`
		Currency     string `yaml:"currency" validate:"required,oneof=RUB EUR USD"`
		Confirmation struct {
			Type      string `yaml:"type" validate:"required"`
			ReturnURL string `yaml:"return_url" validate:"required,url"`
		} `yaml:"confirmation"`
	} `yaml:"creates"`
	TrustedIps []string `yaml:"trusted_ips" validate:"required"`
}

type SMTP struct {
	Host     string `yaml:"host" validate:"required,hostname"`
	Port     int    `yaml:"port" validate:"required,min=1,max=65535"`
	From     string `yaml:"from" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Msg      struct {
		Subject string `yaml:"subject" validate:"required"`
	} `yaml:"msg" validate:"required"`
}

type Middlewares struct {
	RateLimit struct {
		Window time.Duration `yaml:"window" validate:"required,min=1m"`
		Limit  int           `yaml:"limit" validate:"required,gte=1"`
	} `yaml:"rate_limit"`
}

// .env must be loaded
func New(path string) (*Config, error) {
	// try to read the config file
	bytes, err := loadBytes(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// parse the bytes of config file
	cfg, err := parseBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateTLS(cfg *Config) error {
	certPath := filepath.Clean(cfg.Server.HTTP.TLS.ServerCertPath)
	if stat, err := os.Stat(certPath); err != nil || stat.IsDir() {
		return &InvalidCertPathError{
			original: err,
			path:     certPath,
		}
	}

	keyPath := filepath.Clean(cfg.Server.HTTP.TLS.ServerKeyPath)
	if stat, err := os.Stat(keyPath); err != nil || stat.IsDir() {
		return &InvalidKeyPathError{
			original: err,
			path:     keyPath,
		}
	}

	return nil
}

func validateConfig(cfg *Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return err
	}

	// validate tls if it's enable
	if cfg.Server.HTTP.TLS.Enable {
		return validateTLS(cfg)
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
	// validate the path
	path = filepath.Clean(path)
	if _, err := filepath.Abs(path); err != nil {
		return nil, err
	}

	// read the config file
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// load .env variables
	content := os.ExpandEnv(string(bytes))
	return []byte(content), nil
}
