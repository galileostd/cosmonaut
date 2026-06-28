package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// EventType describes what happened.
type EventType string

const (
	EventTypeComponentHealthChanged EventType = "component.health_changed"
	EventTypeComponentCreated       EventType = "component.created"
	EventTypeComponentDeleted       EventType = "component.deleted"
	EventTypeJobStatusChanged       EventType = "job.status_changed"
)

// Event is a real-time notification pushed to WebSocket subscribers.
type Event struct {
	Type      EventType      `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

// EventBus fan-outs events to all connected WebSocket subscribers.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string]chan Event
}

func newEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]chan Event),
	}
}

// Publish sends an event to all current subscribers.
// Non-blocking: subscribers that are slow to consume are dropped.
func (eb *EventBus) Publish(e Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for id, ch := range eb.subscribers {
		select {
		case ch <- e:
		default:
			slog.Warn("event bus: subscriber dropped event (channel full)", "subscriber", id, "event_type", e.Type)
		}
	}
}

func (eb *EventBus) subscribe(id string) chan Event {
	ch := make(chan Event, 64)
	eb.mu.Lock()
	eb.subscribers[id] = ch
	eb.mu.Unlock()
	return ch
}

func (eb *EventBus) unsubscribe(id string) {
	eb.mu.Lock()
	if ch, ok := eb.subscribers[id]; ok {
		close(ch)
		delete(eb.subscribers, id)
	}
	eb.mu.Unlock()
}

// handleEvents upgrades the connection to WebSocket and streams events.
// WS /api/v1/events
//
// Optional query params:
//
//	types: comma-separated list of event types to filter (default: all)
//
// The client receives a JSON stream of Event objects.
// The connection stays open until the client disconnects.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// respect the CORS config already set on the router
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Warn("websocket upgrade failed", "err", err)
		return
	}
	defer conn.CloseNow()

	subID := requestIDFromContext(r.Context())
	ch := s.events.subscribe(subID)
	defer s.events.unsubscribe(subID)

	// send a welcome event so the client knows the connection is live
	welcome := Event{
		Type:      "connected",
		Timestamp: time.Now(),
		Payload:   map[string]any{"message": "connected to Cosmonaut event stream"},
	}
	if err := wsjson.Write(r.Context(), conn, welcome); err != nil {
		return
	}

	// ping loop to keep the connection alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.Ping(r.Context()); err != nil {
					return
				}
			case <-r.Context().Done():
				return
			}
		}
	}()

	// stream events until disconnect
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			if err := wsjson.Write(r.Context(), conn, event); err != nil {
				return
			}
		case <-r.Context().Done():
			conn.Close(websocket.StatusNormalClosure, "")
			return
		}
	}
}

// PublishHealthChanged emits a component.health_changed event.
// Called by the health controller after each reconciliation.
func (eb *EventBus) PublishHealthChanged(namespace, name, health, message string) {
	eb.Publish(Event{
		Type:      EventTypeComponentHealthChanged,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"namespace": namespace,
			"name":      name,
			"health":    health,
			"message":   message,
		},
	})
}

// MarshalJSON implements custom marshaling for Event to handle the
// EventType as a plain string in the top-level "type" field.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  string(e.Type),
		Alias: (Alias)(e),
	})
}

// contextWithCancel is a helper to create a cancellable context tied to a WebSocket conn.
func contextWithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}
