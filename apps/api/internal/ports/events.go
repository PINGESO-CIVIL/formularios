package ports

import (
	"context"
	"encoding/json"
	"time"
)

type Actor struct {
	ID   string
	Role string
}

// Event es un evento de dominio. El sobre definitivo se arma en adapters/events.
// TODO(integracion): broker y formato definitivo del sobre (docs/preguntas-abiertas.md #5).
type Event struct {
	Type          string
	Version       int
	OccurredAt    time.Time
	Actor         Actor
	Context       map[string]string
	CorrelationID string
	Payload       json.RawMessage
}

type EventPublisher interface {
	Publish(ctx context.Context, e Event) error
}
