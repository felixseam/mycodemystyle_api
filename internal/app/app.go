package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"mycodemystyle_api/internal/cache/tarantool"
	"mycodemystyle_api/internal/config"
	"mycodemystyle_api/internal/database/postgres"
	"mycodemystyle_api/internal/logger"
	"mycodemystyle_api/internal/router"
	"mycodemystyle_api/internal/server"
	"mycodemystyle_api/internal/ws"
)

// Run builds the whole application and blocks until it stops.
func Run() error {
	// Init App
	appConfig := config.NewAppConfig()
	logger.Init("info")

	// Init Tarantool cache connection pool
	tConn, err := tarantool.Connect(config.NewTarantoolConfig())
	if err != nil {
		return fmt.Errorf("tarantool initialization failed: %w", err)
	}
	defer tConn.Close()

	cacheClient := tarantool.New(tConn)

	if err := cacheClient.Set(
		context.Background(),
		"app:health",
		map[string]string{"status": "ok"},
		30*time.Second,
	); err != nil {
		return fmt.Errorf("cache write failed: %w", err)
	}
	log.Println("Tarantool cache initialized")

	// Init Postgres database
	pgPool, err := postgres.Connect(config.NewPostgresConfig())
	if err != nil {
		return fmt.Errorf("postgresql initialization failed: %w", err)
	}
	defer pgPool.Close()
	log.Println("Postgres database initialized")

	// Init websocket hub
	frontHub := ws.NewHub()
	adminHub := ws.NewHub()

	// Init Fiber, routes and dependencies
	app := server.New()
	router.New(app, &router.AppDependencies{
		PgDBPool:       pgPool,
		TarantoolCache: cacheClient,
		TimeZone:       appConfig.AppTimezone,
		FrontHub:       frontHub,
		AdminHub:       adminHub,
	}).Register()

	// Serve until Ctrl+C or SIGTERM: hub closes first, then the server stops,
	// then the deferred calls above close Postgres and Tarantool.
	addr := fmt.Sprintf("%s:%d", appConfig.AppHost, appConfig.AppPort)
	return server.Run(app, addr, frontHub.Shutdown, adminHub.Shutdown)
}
