package main

import (
	"log"
	"net/http"
	"os"

	"github.com/femisowemimo/booking-appointment/backend/pkg/bootstrap"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it (using system env)")
	}

	log.Println("Starting Appointment API Service...")

	h := bootstrap.GetHandler()

	// 7. Start Server
	port := os.Getenv("PORT")
	addr := ""

	if port == "" {
		// Local Dev: Listen on localhost only to avoid firewall popup
		addr = "127.0.0.1:8080"
	} else {
		// Prod: Listen on all interfaces
		addr = ":" + port
	}

	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
