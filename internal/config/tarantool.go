package config

import (
	"os"
	"strconv"
	"time"
)

type TarantoolConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}

func NewTarantoolConfig() *TarantoolConfig {
	return &TarantoolConfig{
		Host:     getEnv("TARANTOOL_HOST", "127.0.0.1"),
		Port:     getEnvInt("TARANTOOL_PORT", 3301),
		Username: os.Getenv("TARANTOOL_USER"),
		Password: os.Getenv("TARANTOOL_PASSWORD"),
		Timeout:  5 * time.Second,
	}
}

func (c *TarantoolConfig) Address() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}
