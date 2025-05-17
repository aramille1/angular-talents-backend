package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reverse-job-board/db"
	"reverse-job-board/handlers"
	"reverse-job-board/internal"
	"reverse-job-board/middlewares"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var wg = sync.WaitGroup{}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

// spaHandler implements the http.Handler interface for serving a Single Page Application
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP handles the HTTP request by serving static files or routing to the SPA
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If the request starts with /api or known api endpoints, skip static file handling
	// This allows the Go API handlers to process these requests
	if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/adminski") {
		// Don't do anything - let the API handlers handle this request
		return
	}

	// Check if the request is for a static file
	if strings.Contains(path, ".") {
		// For file requests (with extensions like .js, .css, .png)
		physicalPath := filepath.Join(h.staticPath, path)
		if _, err := os.Stat(physicalPath); !os.IsNotExist(err) {
			// File exists, serve it directly
			http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
			return
		}
	}

	// If we get here, it's either a non-existent file or an Angular route
	// Serve the index.html (SPA fallback)
	indexFile := filepath.Join(h.staticPath, h.indexPath)
	log.Printf("Serving index.html for route: %s", path)
	http.ServeFile(w, r, indexFile)
}

func main() {
	_, error := os.Stat(".env")

	if !os.IsNotExist(error) {
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatalf("Error loading .env file")
		}
	}

	fmt.Println("Create router")
	r := mux.NewRouter()
	r.Use(middlewares.RecoverPanic)

	db.InitiateDB()

	// Create a completely separate adminski router
	adminskiRouter := r.PathPrefix("/adminski").Subrouter()
	// Add admin auth middleware to protect all adminski routes
	adminskiRouter.Use(middlewares.ValidateAdminAuth) // Use admin-specific auth middleware

	// Adminski routes for recruiters
	adminskiRouter.Handle("/recruiters", internal.EnhancedHandler(handlers.HandleRecruiterList)).Methods("GET")
	adminskiRouter.Handle("/recruiters/{recruiterID}/status", internal.EnhancedHandler(handlers.HandleRecruiterUpdateStatus)).Methods("PATCH")

	// Adminski routes for users
	adminskiRouter.Handle("/users/{userID}/email", internal.EnhancedHandler(handlers.HandleGetUserEmail)).Methods("GET")

	// Regular API routes
	r.Handle("/health", internal.EnhancedHandler(handlers.HandleHealth)).Methods("GET")
	r.Handle("/email", internal.EnhancedHandler(handlers.HandleEmail)).Methods("GET")
	r.Handle("/sign-up", internal.EnhancedHandler(handlers.HandleSignUp)).Methods("POST")
	r.Handle("/login", internal.EnhancedHandler(handlers.HandleLogin)).Methods("POST")
	r.Handle("/verify/{userID}/{verificationCode}", internal.EnhancedHandler(handlers.HandleEmailVerify)).Methods("GET")
	r.Handle("/count", internal.EnhancedHandler(handlers.HandleCount)).Methods("GET")

	// Admin login route
	r.Handle("/api/admin/login", internal.EnhancedHandler(handlers.HandleAdminLogin)).Methods("POST")

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

	// Get the static files directory from environment or use default
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "../angular-dist"  // Default to a relative path
	}

	// Create a spa handler for the Angular app
	spa := spaHandler{staticPath: staticDir, indexPath: "index.html"}

	// Use the spa handler for all routes not matched by the API routes
	r.PathPrefix("/").Handler(spa)

	withCors := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://angulartalents.onrender.com", "https://www.angulartalents.com", "http://localhost:4200"},
		AllowedMethods:   []string{"GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Access-Control-Allow-Headers", "Origin", "Accept", "X-Requested-With", "Content-Type", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Listening to port %s\n", port)
	http.ListenAndServe(":"+port, withCors)
}
