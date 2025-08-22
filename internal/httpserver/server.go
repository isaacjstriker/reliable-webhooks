package httpserver

import (
	"github.com/gofiber/fiber/v2"
	"github.com/isaacjstriker/reliable-webhooks/internal/config"
	"github.com/isaacjstriker/reliable-webhooks/internal/httpserver/handlers"
	"github.com/isaacjstriker/reliable-webhooks/internal/processing"
	"github.com/isaacjstriker/reliable-webhooks/internal/repository"
	"log/slog"
	"time"
)

func New(cfg config.Config, logger *slog.Logger, repo repository.EventRepository, dispatcher *processing.Dispatcher) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		AppName:      "reliable-webhooks",
	})

	app.Use(requestLogger(logger))

	app.Get("/healthz", handlers.Health)
	app.Get("/readyz", handlers.Ready(repo))

	app.Post("/webhooks/stripe", handlers.StripeWebhook(logger, repo, dispatcher))

	return app
}

func requestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Info("http_request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}
}