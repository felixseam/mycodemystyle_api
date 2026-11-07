package config

import (
	"os"
)

type PostgresConfig struct {
	DatabaseURL string
	MaxConns    int32
	MinConns    int32
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		DatabaseURL: os.Getenv("PG_DATABASE_URL"),
		MaxConns:    int32(getEnvInt("PG_MAX_CONNS", 20)),
		MinConns:    int32(getEnvInt("PG_MIN_CONNS", 2)),
	}
}
