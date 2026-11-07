package tarantool

import (
	"context"
	"fmt"
	"mycodemystyle_api/internal/config"

	"github.com/tarantool/go-tarantool/v2"
)

func Connect(cfg *config.TarantoolConfig) (*tarantool.Connection, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Timeout,
	)
	defer cancel()

	dialer := tarantool.NetDialer{
		Address:  cfg.Address(),
		User:     cfg.Username,
		Password: cfg.Password,
	}

	conn, err := tarantool.Connect(ctx, dialer, tarantool.Opts{
		Timeout:     cfg.Timeout,
		Concurrency: 8,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to Tarantool: %w", err)
	}

	return conn, nil
}
