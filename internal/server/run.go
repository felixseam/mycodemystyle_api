package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

// Run starts the HTTP server and blocks until it fails or a stop signal
// (Ctrl+C, docker stop) arrives. The beforeShutdown functions run first,
// for example hub.Shutdown, then the server stops.
func Run(app *fiber.App, addr string, beforeShutdown ...func()) error {
	errCh := make(chan error, 1)
	go func() { errCh <- app.Listen(addr) }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		log.Printf("shutting down (%v)", sig)
	}

	for _, fn := range beforeShutdown {
		fn()
	}
	return app.Shutdown()
}
