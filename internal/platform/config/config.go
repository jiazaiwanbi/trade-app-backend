package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	defaultAppEnv         = "local"
	defaultHTTPAddr       = ":8080"
	defaultHTTPTimeoutSec = 15
	defaultPostgresHost   = "localhost"
	defaultPostgresPort   = 5432
	defaultPostgresDB     = "trade_app"
	defaultPostgresUser   = "postgres"
	defaultPostgresPass   = "postgres"
	defaultPostgresSSL    = "disable"
	defaultJWTExpireHours = 24
)

type Config struct {
	AppEnv       string
	HTTPAddr     string
	HTTPTimeout  time.Duration
	ShutdownWait time.Duration
	JWT          JWTConfig
	Postgres     PostgresConfig
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

func Load() (Config, error) {
	cfg, err := loadBase()
	if err != nil {
		return Config{}, err
	}

	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func LoadForMigrate() (Config, error) {
	return loadBase()
}

func loadBase() (Config, error) {
	httpTimeoutSec, err := getEnvAsInt("APP_HTTP_TIMEOUT_SECONDS", defaultHTTPTimeoutSec)
	if err != nil {
		return Config{}, fmt.Errorf("load APP_HTTP_TIMEOUT_SECONDS: %w", err)
	}
	if httpTimeoutSec <= 0 {
		return Config{}, fmt.Errorf("APP_HTTP_TIMEOUT_SECONDS must be greater than 0")
	}

	jwtExpireHours, err := getEnvAsInt("JWT_EXPIRE_HOURS", defaultJWTExpireHours)
	if err != nil {
		return Config{}, fmt.Errorf("load JWT_EXPIRE_HOURS: %w", err)
	}
	if jwtExpireHours <= 0 {
		return Config{}, fmt.Errorf("JWT_EXPIRE_HOURS must be greater than 0")
	}

	postgresPort, err := getEnvAsInt("POSTGRES_PORT", defaultPostgresPort)
	if err != nil {
		return Config{}, fmt.Errorf("load POSTGRES_PORT: %w", err)
	}
	if postgresPort <= 0 {
		return Config{}, fmt.Errorf("POSTGRES_PORT must be greater than 0")
	}

	return Config{
		AppEnv:       getEnv("APP_ENV", defaultAppEnv),
		HTTPAddr:     getEnv("APP_HTTP_ADDR", defaultHTTPAddr),
		HTTPTimeout:  time.Duration(httpTimeoutSec) * time.Second,
		ShutdownWait: 10 * time.Second,
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			TTL:    time.Duration(jwtExpireHours) * time.Hour,
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", defaultPostgresHost),
			Port:     postgresPort,
			Database: getEnv("POSTGRES_DB", defaultPostgresDB),
			User:     getEnv("POSTGRES_USER", defaultPostgresUser),
			Password: getEnv("POSTGRES_PASSWORD", defaultPostgresPass),
			SSLMode:  getEnv("POSTGRES_SSLMODE", defaultPostgresSSL),
		},
	}, nil
}

func (p PostgresConfig) URL() string {
	values := url.Values{}
	values.Set("sslmode", p.SSLMode)

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?%s",
		url.QueryEscape(p.User),
		url.QueryEscape(p.Password),
		p.Host,
		p.Port,
		p.Database,
		values.Encode(),
	)
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) (int, error) {
	value := getEnv(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as int: %w", key, err)
	}

	return parsed, nil
}
