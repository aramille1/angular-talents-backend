package main

import (
	"angular-talents-backend/db"
	"angular-talents-backend/handlers"
	"angular-talents-backend/internal"
	"angular-talents-backend/middlewares"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var wg = sync.WaitGroup{}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {
	fmt.Println("Create router")
	r := mux.NewRouter()
	r.Use(middlewares.RecoverPanic)

	db.InitiateDB()

	r.Handle("/sign-up", internal.EnhancedHandler(handlers.HandleSignUp)).Methods("POST")
	r.Handle("/login", internal.EnhancedHandler(handlers.HandleLogin)).Methods("POST")

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

	fmt.Println("Listening to port 3000")
	http.ListenAndServe(":3000", withCors)
}
