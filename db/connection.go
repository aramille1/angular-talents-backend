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
)

var Database *mongo.Database

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

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions)

	clientOptions.SetServerSelectionTimeout(10 * time.Second)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Println("Failed to create new client")
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		fmt.Println("Failed to connect to client")
		log.Fatal(err)
	}

	// defer func() {
	//     if err = client.Disconnect(ctx); err != nil {
	//         panic(err)
	//     }
	// }()

	Database = client.Database(dbName)

	fmt.Println("Pinging db!")
	if err := client.Ping(ctx, nil); err != nil {
		fmt.Println("Failed to ping database:", err)
		fmt.Println("Continuing despite ping failure - DB operations may fail")
		return
	}

	fmt.Println("Successfully connected database")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
