package bootstrap

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/handlers"
	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/messaging"
	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/repositories"
	"github.com/femisowemimo/booking-appointment/backend/pkg/core/services"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	Repo      *repositories.PostgresAppointmentRepository
	Publisher *messaging.RabbitMQPublisher
	server    http.Handler
	once      sync.Once
)

func GetHandler() http.Handler {
	once.Do(func() {
		log.Println("Initializing Appointment API Service...")

		// 1. Database Connection
		dbConnStr := os.Getenv("DATABASE_URL")
		if dbConnStr == "" {
			log.Println("WARNING: DATABASE_URL is not set. Defaulting to localhost.")
			dbConnStr = "postgres://user:password@localhost:5432/appointments?sslmode=disable"
		} else {
			log.Println("Found DATABASE_URL, attempting connection...")
		}

		// Use a recovery middleware for the entire init process in case of early panics
		// But here we just want to avoid os.Exit/log.Fatal
		db, err := sql.Open("postgres", dbConnStr)
		if err != nil {
			log.Printf("ERROR: Failed to open DB driver: %v", err)
			// We continue, but later requests might fail if db is nil or invalid.
			// sql.Open usually doesn't fail unless driver is missing.
		}

		// 1.5 Run Migrations
		// Only attempt if we have a DB object
		if db != nil {
			paths := []string{
				"backend/migrations/001_initial_schema.sql", // From repo root
				"migrations/001_initial_schema.sql",         // From backend root
				"../migrations/001_initial_schema.sql",      // From api/index.go relative path
				"./migrations/001_initial_schema.sql",       // Local relative
			}

			var migrationBytes []byte
			var readErr error

			for _, path := range paths {
				migrationBytes, readErr = os.ReadFile(path)
				if readErr == nil {
					log.Printf("Found migration file at: %s", path)
					break
				}
			}

			if readErr != nil {
				log.Printf("Warning: Could not read migration file: %v", readErr)
			} else {
				if _, err := db.Exec(string(migrationBytes)); err != nil {
					log.Printf("Warning: Failed to run migrations: %v", err)
				} else {
					log.Println("Database migrations applied successfully.")
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				log.Printf("Warning: DB ping failed or timed out: %v", err)
			}
		}

		// 2. RabbitMQ Connection
		amqpConnStr := os.Getenv("RABBITMQ_URL")
		var rabbitConn *amqp.Connection
		if amqpConnStr != "" {
			rabbitConn, err = amqp.Dial(amqpConnStr)
			if err != nil {
				log.Printf("Warning: Failed to connect to RabbitMQ: %v", err)
			}
		}

		// 3. Initialize Adapters
		if db != nil {
			Repo = repositories.NewPostgresAppointmentRepository(db)
		} else {
			// Mock or nil repo to prevent nil pointer panics in service?
			// Better: let it be nil, but handle nil in service?
			// For now, we assume NewPostgresAppointmentRepository handles or we cope.
			// Actually repositories.NewPostgresAppointmentRepository(nil) might be fine until Method call.
		}

		// Handle optional publisher
		if rabbitConn != nil {
			Publisher, err = messaging.NewRabbitMQPublisher(rabbitConn)
			if err != nil {
				log.Printf("Warning: Failed to init publisher: %v", err)
			}
		} else {
			log.Println("Running without messaging publisher")
		}

		// 4. Initialize Core Service
		// Check for nil Repo if needed, but assuming Service handles it or fail fast on request
		svc := services.NewAppointmentService(Repo, Publisher)

		// 5. Initialize Handlers
		h := handlers.NewAppointmentHandler(svc)

		// 6. Routes
		mux := http.NewServeMux()

		healthHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}

		appointmentHandler := func(w http.ResponseWriter, r *http.Request) {
			// Double check: if Repo is nil, return 500 error cleanly
			if Repo == nil {
				http.Error(w, "Database connection unavailable", http.StatusServiceUnavailable)
				return
			}

			if r.Method == http.MethodPost {
				h.Create(w, r)
			} else if r.Method == http.MethodGet {
				h.Get(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}

		// Handle both root paths and /api prefixed paths for Vercel/Local compatibility
		// When Vercel rewrites /api/foo -> dest, req.URL.Path might still be /api/foo

		mux.HandleFunc("/health", healthHandler)
		mux.HandleFunc("/api/health", healthHandler)

		mux.HandleFunc("/appointments", appointmentHandler)
		mux.HandleFunc("/api/appointments", appointmentHandler)

		// Catch-all to see what's happening with unmatched routes
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("DEBUG: Unmatched route: %s %s", r.Method, r.URL.Path)
			http.Error(w, "Not Found (Catch-All)", http.StatusNotFound)
		})

		server = enableCORS(mux)
	})
	return server
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log every request to Vercel logs
		log.Printf("DEBUG: Request received: %s %s RemoteAddr: %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Set CORS headers for ALL responses
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
