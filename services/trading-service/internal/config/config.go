package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	EngineHost        string
	EnginePort        int
	HeartbeatInterval time.Duration
	ProtocolVersion   uint32
	HTTPPort          int

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	TransfermarktAPIURL string
	FrontendOrigin      string
}

func Load() Config {
	return Config{
		EngineHost:        getEnv("ENGINE_HOST", "matching-engine"),
		EnginePort:        getEnvInt("ENGINE_PORT", 9000),
		HeartbeatInterval: getEnvDuration("HEARTBEAT_INTERVAL", 5*time.Second),
		ProtocolVersion:   uint32(getEnvInt("PROTOCOL_VERSION", 1)),
		HTTPPort:          getEnvInt("HTTP_PORT", 8080),

		DBHost:     getEnv("DB_HOST", "postgres"),
		DBPort:     getEnvInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "trading"),
		DBPassword: getEnv("DB_PASSWORD", "trading"),
		DBName:     getEnv("DB_NAME", "trading"),

		TransfermarktAPIURL: getEnv("TRANSFERMARKT_API_URL", "http://transfermarkt-api:8000"),
		FrontendOrigin:      getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
	}
}

// DSN builds a Postgres connection string from the DB_* fields.
func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// MigrateDSN is the same connection, using the scheme golang-migrate's
// pgx/v5 driver registers itself under.
func (c Config) MigrateDSN() string {
	return fmt.Sprintf("pgx5://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
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

func getEnvDuration(key string, def time.Duration) time.Duration {
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
