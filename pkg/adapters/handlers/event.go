package handlers

import (
	"encoding/json"
	"net/http"
)

type Event struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Venue string `json:"venue"`
}

type EventHandler struct{}

func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	// Matching data from TicketBookingFlow.tsx
	events := []Event{
		{ID: "event-1", Name: "Late Night Comedy", Venue: "The Basement Club"},
		{ID: "event-2", Name: "Jazz Quartet", Venue: "Blue Note Lounge"},
		{ID: "event-3", Name: "Indie Film Festival", Venue: "Cinema 4"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
