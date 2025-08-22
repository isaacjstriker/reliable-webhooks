package handlers

import "github.com/gofiber/fiber/v2"

func Health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

func Ready(repoReady interface{ Ping() error }) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := repoReady.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready", "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	}
}