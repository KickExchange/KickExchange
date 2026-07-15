package config

import (
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
}

func Load() Config {
	return Config{
		EngineHost:        getEnv("ENGINE_HOST", "matching-engine"),
		EnginePort:        getEnvInt("ENGINE_PORT", 9000),
		HeartbeatInterval: getEnvDuration("HEARTBEAT_INTERVAL", 5*time.Second),
		ProtocolVersion:   uint32(getEnvInt("PROTOCOL_VERSION", 1)),
		HTTPPort:          getEnvInt("HTTP_PORT", 8080),
	}
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
