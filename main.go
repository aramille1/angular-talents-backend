package main

import (
	"angular-talents-backend/db"
	"angular-talents-backend/handlers"
	"angular-talents-backend/internal"
	"angular-talents-backend/middlewares"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var wg = sync.WaitGroup{}
var startTime time.Time

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	startTime = time.Now()
}

func main() {
	fmt.Println("Create router")
	r := mux.NewRouter()
	r.Use(middlewares.RecoverPanic)

	// Add health check endpoint
	r.HandleFunc("/health", handleHealthCheck).Methods("GET")

	db.InitiateDB()

	r.Handle("/health", internal.EnhancedHandler(handlers.HandleHealth)).Methods("GET")
	r.Handle("/email", internal.EnhancedHandler(handlers.HandleEmail)).Methods("GET")
	r.Handle("/sign-up", internal.EnhancedHandler(handlers.HandleSignUp)).Methods("POST")
  r.Handle("/login", internal.EnhancedHandler(handlers.HandleLogin)).Methods("POST")
	r.Handle("/verify/{userID}/{verificationCode}", internal.EnhancedHandler(handlers.HandleEmailVerify)).Methods("GET")
  r.Handle("/count", internal.EnhancedHandler(handlers.HandleCount)).Methods("GET")

	authenticatedRoutes := r.NewRoute().Subrouter()

	authenticatedRoutes.Use(middlewares.ValidateAuth)
	authenticatedRoutes.Handle("/me", internal.EnhancedHandler(handlers.HandleAuthenticatedUserRead)).Methods("GET")

	authenticatedRoutes.Handle("/engineers/me", internal.EnhancedHandler(handlers.HandleAuthenticatedEngineerUpdate)).Methods("PUT")
	authenticatedRoutes.Handle("/engineers", internal.EnhancedHandler(handlers.HandleEngineerCreate)).Methods("POST")

	authenticatedRoutes.Handle("/recruiters/me", internal.EnhancedHandler(handlers.HandleAuthenticatedRecruiterUpdate)).Methods("PUT")
	authenticatedRoutes.Handle("/recruiters", internal.EnhancedHandler(handlers.HandleRecruiterCreate)).Methods("POST")

	membersRoutes := r.NewRoute().Subrouter()

	membersRoutes.Use(middlewares.ValidateMembership)
	membersRoutes.Handle("/engineers", internal.EnhancedHandler(handlers.HandleEngineerList)).Methods("GET")
	membersRoutes.Handle("/engineers/{engineerID}", internal.EnhancedHandler(handlers.HandleEngineerRead)).Methods("GET")

	withCors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "HEAD", "OPTIONS", "POST", "PUT"},
		AllowedHeaders:   []string{"Authorization", "Access-Control-Allow-Headers", "Origin", "Accept", "X-Requested-With", "Content-Type", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug:                true,
		OptionsSuccessStatus: 200,
	}).Handler(r)

	// Get port from environment variable, default to 3000
	port := getEnv("PORT", "3000")

	fmt.Printf("Listening on port %s\n", port)
	http.ListenAndServe(":"+port, withCors)
}

// Health check handler for monitoring
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Create health check response
	response := map[string]interface{}{
		"status":      "ok",
		"uptime":      time.Since(startTime).String(),
		"version":     "1.0.0",
		"environment": getEnv("ENVIRONMENT", "development"),
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Write response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper function to get environment variables with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
