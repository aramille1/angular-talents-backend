package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Admin represents the admin document in MongoDB
type Admin struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Username      string             `bson:"username"`
	Password      string             `bson:"password"`
	Email         string             `bson:"email"`
	FirstName     string             `bson:"firstName"`
	LastName      string             `bson:"lastName"`
	IsSuper       bool               `bson:"isSuper"`
	AdminVerified bool               `bson:"adminVerified"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file. Using environment variables.")
	}

	// Get MongoDB connection details from environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Get database name from environment or use default
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "angular-talents"
	}

	// Access the admin collection
	db := client.Database(dbName)
	collection := db.Collection("admin")

	// Check if admin already exists
	username := "admin"
	var existingAdmin Admin
	err = collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&existingAdmin)
	if err == nil {
		log.Printf("Admin user '%s' already exists", username)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Fatalf("Error checking for existing admin: %v", err)
	}

	// Hash the password
	password := "adminpass123" // You should change this to a secure password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create the admin document
	admin := Admin{
		Username:      username,
		Password:      string(hashedPassword),
		Email:         "admin@example.com",
		FirstName:     "Admin",
		LastName:      "User",
		IsSuper:       true,
		AdminVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert the admin document
	_, err = collection.InsertOne(context.Background(), admin)
	if err != nil {
		log.Fatalf("Failed to insert admin: %v", err)
	}

	fmt.Println("Admin user created successfully!")
	fmt.Println("Username:", username)
	fmt.Println("Password:", password)
}
