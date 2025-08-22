package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/isaacjstriker/reliable-webhooks/internal/domain"
	"github.com/isaacjstriker/reliable-webhooks/internal/processing"
	"github.com/isaacjstriker/reliable-webhooks/internal/repository"
	"log/slog"
	"strings"
)

type stripeEvent struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// NOTE: Real implementation must verify signature using Stripe's library or manual HMAC.
// This is a placeholder for early vertical slice.
func StripeWebhook(logger *slog.Logger, repo repository.EventRepository, dispatcher *processing.Dispatcher) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sig := c.Get("Stripe-Signature")
		if sig == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing signature"})
		}

		body := c.Body()
		var evt stripeEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_json"})
		}
		if strings.TrimSpace(evt.ID) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing_id"})
		}

		// TODO: Verify signature (iteration 2)

		event := domain.Event{
			Provider: "stripe",
			EventID:  evt.ID,
			Payload:  body,
		}

		err := repo.InsertIfNew(c.Context(), &event)
		if err != nil {
			if errors.Is(err, repository.ErrDuplicateEvent) {
				logger.Info("duplicate_event", "provider", event.Provider, "event_id", event.EventID)
				return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "duplicate_ignored"})
			}
			logger.Error("event_insert_error", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal"})
		}

		// Enqueue for processing
		select {
		case dispatcher.Enqueue(event.ID):
		default:
			logger.Error("dispatcher_queue_full", "event_id", event.ID)
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "received"})
	}
}