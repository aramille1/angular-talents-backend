package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	Database *mongo.Database
	Client   *mongo.Client
)

func InitiateDB() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables or defaults")
	}

	// Get MongoDB connection parameters from environment variables
	username := getEnv("MONGODB_USERNAME", "aramille")
	password := getEnv("MONGODB_PASSWORD", "")
	cluster := getEnv("MONGODB_CLUSTER", "atdbcluster0.3ynluj2.mongodb.net")
	dbName := getEnv("MONGODB_DATABASE", "ATDB-cluster")

	var uri string

	// If we have MongoDB Atlas credentials, use them
	if username != "" && password != "" && cluster != "" {
		uri = fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=ATDBCluster0",
			username,
			password,
			cluster)
		fmt.Println("Connecting to MongoDB Atlas")
	} else {
		// Fallback to local MongoDB
		uri = "mongodb://localhost:27017"
		fmt.Println("Connecting to local MongoDB")
	}

	// Configure MongoDB client with retry options
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetServerSelectionTimeout(15 * time.Second).
		SetConnectTimeout(20 * time.Second)

	// Create a new client and connect to MongoDB
	maxConnectionAttempts := 5
	var err error
	var client *mongo.Client

	for attempt := 1; attempt <= maxConnectionAttempts; attempt++ {
		fmt.Printf("MongoDB connection attempt %d of %d\n", attempt, maxConnectionAttempts)
		client, err = mongo.NewClient(clientOptions)
		if err != nil {
			fmt.Printf("Failed to create MongoDB client (attempt %d/%d): %v\n", attempt, maxConnectionAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err = client.Connect(ctx)

		if err != nil {
			cancel()
			fmt.Printf("Failed to connect to MongoDB (attempt %d/%d): %v\n", attempt, maxConnectionAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Test the connection with a ping
		err = client.Ping(ctx, readpref.Primary())
		cancel()

		if err != nil {
			fmt.Printf("Failed to ping MongoDB (attempt %d/%d): %v\n", attempt, maxConnectionAttempts, err)
			// Close the connection before retrying
			if client != nil {
				_ = client.Disconnect(context.Background())
			}
			time.Sleep(2 * time.Second)
			continue
		}

		// Connection successful
		Client = client
		Database = client.Database(dbName)
		fmt.Println("Successfully connected to MongoDB")

		// Register disconnect function when application exits
		// This helps with clean shutdown
		// runtime.SetFinalizer(&client, func(c **mongo.Client) {
		// 	(*c).Disconnect(context.Background())
		// })

		return
	}

	// If we've reached here, all connection attempts failed
	if err != nil {
		fmt.Printf("Failed to connect to MongoDB after %d attempts: %v\n", maxConnectionAttempts, err)

		// For production environments, we might want to continue despite failed DB connection
		// But for development, it's better to fail early
		if getEnv("ENVIRONMENT", "development") == "development" {
			log.Fatalf("Failed to connect to MongoDB after %d attempts: %v\n", maxConnectionAttempts, err)
		} else {
			fmt.Println("WARNING: Continuing without MongoDB connection. Application functionality will be limited.")
		}
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
