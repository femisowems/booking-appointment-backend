package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/ports"
)

type AppointmentHandler struct {
	service ports.AppointmentService
}

func NewAppointmentHandler(service ports.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{service: service}
}

type CreateAppointmentRequest struct {
	UserID     string    `json:"user_id"`
	ProviderID string    `json:"provider_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

func (h *AppointmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	appt, err := h.service.Create(r.Context(), req.UserID, req.ProviderID, req.StartTime, req.EndTime)
	if err != nil {
		// Differentiate errors (client vs server)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(appt)
}

func (h *AppointmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Check for provider_id query param
	providerID := r.URL.Query().Get("provider_id")
	if providerID != "" {
	// List by provider with date range
		startStr := r.URL.Query().Get("start_date")
		endStr := r.URL.Query().Get("end_date")
		
		now := time.Now()
		// Default to today
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)

		if startStr != "" {
			if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
				start = parsed
			} else {
				// Log ignored error
				json.NewEncoder(w).Encode(map[string]string{"error": "parse error start", "details": err.Error(), "input": startStr})
				return
			}
		}
		if endStr != "" {
			if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
				end = parsed
			} else {
				json.NewEncoder(w).Encode(map[string]string{"error": "parse error end", "details": err.Error(), "input": endStr})
				return
			}
		}

		// DEBUG LOG
		// log.Printf("Querying Provider: %s, Start: %v, End: %v", providerID, start, end)
		
		appts, err := h.service.ListByProvider(r.Context(), providerID, start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(appts)
		return
	}

	// Get by ID
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id or provider_id", http.StatusBadRequest)
		return
	}

	appt, err := h.service.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if appt == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(appt)
}
