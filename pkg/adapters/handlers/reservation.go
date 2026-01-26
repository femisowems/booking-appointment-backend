package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/ports"
)

type ReservationHandler struct {
	service ports.ReservationService
}

func NewReservationHandler(service ports.ReservationService) *ReservationHandler {
	return &ReservationHandler{service: service}
}

type CreateReservationRequest struct {
	UserID      string    `json:"user_id"`
	EventID     string    `json:"event_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TicketCount int       `json:"ticket_count"`
}

func (h *ReservationHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default to 1 ticket if not specified
	if req.TicketCount <= 0 {
		req.TicketCount = 1
	}

	res, err := h.service.Create(r.Context(), req.UserID, req.EventID, req.StartTime, req.EndTime, req.TicketCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *ReservationHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Check for event_id query param
	eventID := r.URL.Query().Get("event_id")
	if eventID != "" {
		// List by event with date range
		startStr := r.URL.Query().Get("start_date")
		endStr := r.URL.Query().Get("end_date")

		now := time.Now()
		// Default to today
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)

		if startStr != "" {
			if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
				start = parsed
			}
		}
		if endStr != "" {
			if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
				end = parsed
			}
		}

		res, err := h.service.ListByEvent(r.Context(), eventID, start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	// Get by ID
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id or event_id", http.StatusBadRequest)
		return
	}

	res, err := h.service.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(res)
}
