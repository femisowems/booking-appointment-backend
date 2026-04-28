package bootstrap

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/handlers"
	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/messaging"
	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/repositories"
	"github.com/femisowemimo/booking-appointment/backend/pkg/core/services"
	"github.com/femisowemimo/booking-appointment/backend/pkg/core/ports"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	Repo      ports.ReservationRepository
	Publisher *messaging.RabbitMQPublisher
	server    http.Handler
	once      sync.Once
)

func GetHandler() http.Handler {
	once.Do(func() {
		log.Println("Initializing Reservation API Service...")

		// 1. Database Connection
		dbConnStr := os.Getenv("DATABASE_URL")
		if dbConnStr == "" {
			log.Println("WARNING: DATABASE_URL is not set. Defaulting to localhost.")
			dbConnStr = "postgres://user:password@localhost:5432/reservations?sslmode=disable"
		} else {
			log.Println("Found DATABASE_URL, attempting connection...")
		}

		db, err := connectPostgres(dbConnStr)
		if err != nil {
			log.Printf("Warning: Failed to connect to database: %v", err)
		}

		// 1.5 Run Migrations
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
				log.Printf("Warning: DB ping failed or timed out after migrations: %v", err)
				if closeErr := db.Close(); closeErr != nil {
					log.Printf("Warning: failed to close unhealthy DB connection: %v", closeErr)
				}
				db = nil
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
			Repo = repositories.NewPostgresReservationRepository(db)
		} else {
			log.Println("Warning: using in-memory reservation repository; data will not persist across restarts")
			Repo = repositories.NewInMemoryReservationRepository()
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
		svc := services.NewReservationService(Repo, Publisher)

		// 5. Initialize Handlers
		h := handlers.NewReservationHandler(svc)

		// 6. Routes
		mux := http.NewServeMux()

		healthHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}

		reservationHandler := func(w http.ResponseWriter, r *http.Request) {
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

		mux.HandleFunc("/health", healthHandler)
		mux.HandleFunc("/api/health", healthHandler)

		mux.HandleFunc("/reservations", reservationHandler)
		mux.HandleFunc("/api/reservations", reservationHandler)

		eventHandler := handlers.NewEventHandler()
		mux.HandleFunc("/events", eventHandler.List)
		mux.HandleFunc("/api/events", eventHandler.List)

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("DEBUG: Unmatched route: %s %s", r.Method, r.URL.Path)
			http.Error(w, "Not Found (Catch-All)", http.StatusNotFound)
		})

		server = enableCORS(mux)
	})
	return server
}

func connectPostgres(dbConnStr string) (*sql.DB, error) {
	candidates := []string{dbConnStr}

	if fallbackConn, ok := withSSLModeRequire(dbConnStr); ok {
		candidates = append(candidates, fallbackConn)
	}

	var lastErr error
	for index, candidate := range candidates {
		db, err := sql.Open("postgres", candidate)
		if err != nil {
			lastErr = err
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr := db.PingContext(ctx)
		cancel()
		if pingErr != nil {
			lastErr = pingErr
			if closeErr := db.Close(); closeErr != nil {
				log.Printf("Warning: failed to close DB connection after failed ping: %v", closeErr)
			}
			if index == 0 && len(candidates) > 1 {
				log.Printf("Warning: initial DB connection failed (%v), retrying with sslmode=require", pingErr)
			}
			continue
		}

		if index == 1 {
			log.Println("Connected to database using sslmode=require fallback")
		}
		return db, nil
	}

	return nil, lastErr
}

func withSSLModeRequire(dbConnStr string) (string, bool) {
	u, err := url.Parse(dbConnStr)
	if err != nil {
		return "", false
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", false
	}

	host := strings.Split(u.Hostname(), ":")[0]
	if host == "" {
		return "", false
	}

	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return "", false
	}

	query := u.Query()
	if query.Get("sslmode") != "" {
		return "", false
	}

	query.Set("sslmode", "require")
	u.RawQuery = query.Encode()
	return u.String(), true
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
