package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envAppEnv             = "APP_ENV"
	envHTTPAddr           = "HTTP_ADDR"
	envHTTPReadTimeout    = "HTTP_READ_TIMEOUT"
	envHTTPWriteTimeout   = "HTTP_WRITE_TIMEOUT"
	envHTTPIdleTimeout    = "HTTP_IDLE_TIMEOUT"
	envDBDSN              = "DB_DSN"
	envDBMaxConns         = "DB_MAX_CONNS"
	envDBMinConns         = "DB_MIN_CONNS"
	envDBMaxConnLifetime  = "DB_MAX_CONN_LIFETIME"
	envDBMaxConnIdleTime  = "DB_MAX_CONN_IDLE_TIME"
	envLogLevel           = "LOG_LEVEL"
	envFeatureEnableStats = "FEATURE_ENABLE_STATS"
	envShutdownTimeout    = "SHUTDOWN_TIMEOUT"
)

const (
	defaultEnv                = "local"
	defaultHTTPAddr           = ":8080"
	defaultHTTPReadTimeout    = 10 * time.Second
	defaultHTTPWriteTimeout   = 10 * time.Second
	defaultHTTPIdleTimeout    = 60 * time.Second
	defaultDBMaxConns         = 10
	defaultDBMinConns         = 0
	defaultDBMaxConnLifetime  = 30 * time.Minute
	defaultDBMaxConnIdleTime  = 5 * time.Minute
	defaultLogLevel           = "info"
	defaultFeatureEnableStats = false
	defaultShutdownTimeout    = 5 * time.Second
)

var (
	ErrDBDSNRequired = errors.New("DB_DSN is required")
)

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DBConfig struct {
	DSN             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type LogConfig struct {
	Level string
}

type FeaturesConfig struct {
	EnableStats bool
}

type Config struct {
	Env             string
	HTTP            HTTPConfig
	DB              DBConfig
	Log             LogConfig
	Features        FeaturesConfig
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env: envOrDefault(envAppEnv, defaultEnv),

		HTTP: HTTPConfig{
			Addr:         envOrDefault(envHTTPAddr, defaultHTTPAddr),
			ReadTimeout:  durationOrDefault(envHTTPReadTimeout, defaultHTTPReadTimeout),
			WriteTimeout: durationOrDefault(envHTTPWriteTimeout, defaultHTTPWriteTimeout),
			IdleTimeout:  durationOrDefault(envHTTPIdleTimeout, defaultHTTPIdleTimeout),
		},

		DB: DBConfig{
			DSN:             os.Getenv(envDBDSN),
			MaxConns:        int32(intOrDefault(envDBMaxConns, defaultDBMaxConns)),
			MinConns:        int32(intOrDefault(envDBMinConns, defaultDBMinConns)),
			MaxConnLifetime: durationOrDefault(envDBMaxConnLifetime, defaultDBMaxConnLifetime),
			MaxConnIdleTime: durationOrDefault(envDBMaxConnIdleTime, defaultDBMaxConnIdleTime),
		},

		Log: LogConfig{
			Level: envOrDefault(envLogLevel, defaultLogLevel),
		},

		Features: FeaturesConfig{
			EnableStats: boolOrDefault(envFeatureEnableStats, defaultFeatureEnableStats),
		},

		ShutdownTimeout: durationOrDefault(envShutdownTimeout, defaultShutdownTimeout),
	}

	if cfg.DB.DSN == "" {
		return Config{}, ErrDBDSNRequired
	}

	cfg.Log.Level = strings.ToLower(cfg.Log.Level)

	return cfg, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durationOrDefault(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func intOrDefault(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func boolOrDefault(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
