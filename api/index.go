package handler

import (
	"net/http"
	"runtime/debug"

	"github.com/femisowemimo/booking-appointment/backend/pkg/bootstrap"
)

// Handler is the entrypoint for Vercel Serverless Functions
func Handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			// Ensure CORS headers are set even in case of panic
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")

			http.Error(w, "Internal Server Error: Panic detected", http.StatusInternalServerError)
			// Log the panic for Vercel logs
			// In a real app, you might want to print the stack trace
			println("PANIC RECOVERED:", err)
			debug.PrintStack()
		}
	}()

	h := bootstrap.GetHandler()
	h.ServeHTTP(w, r)
}
