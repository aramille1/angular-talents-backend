package main

import (
	"fmt"
	"net/http"
	"os"
	"reverse-job-board/db"
	"reverse-job-board/handlers"
	"reverse-job-board/internal"
	"reverse-job-board/middlewares"
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
		AllowedOrigins:   []string{"https://angulartalents.onrender.com", "https://www.angulartalents.com", "http://localhost:4200"},
		AllowedMethods:   []string{"GET", "HEAD", "OPTIONS", "POST", "PUT"},
		AllowedHeaders:   []string{"Authorization", "Access-Control-Allow-Headers", "Origin", "Accept", "X-Requested-With", "Content-Type", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler(r)

	fmt.Println("Listening to port 8080")
	http.ListenAndServe(":8080", withCors)
}
