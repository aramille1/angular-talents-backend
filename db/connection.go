package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Database *mongo.Database

func InitiateDB() {
	// Check if local MongoDB should be used
	useLocalDB := os.Getenv("USE_LOCAL_DB")

	var clientOptions *options.ClientOptions
	var dbName string

	if useLocalDB == "true" {
		// Use local MongoDB for development/testing
		fmt.Println("Using local MongoDB instance")
		uri := "mongodb://localhost:27017"
		clientOptions = options.Client().ApplyURI(uri)
		dbName = "angular-talents"
	} else {
		// Use MongoDB Atlas for production
		username := getEnv("MONGODB_USERNAME", "aramille")
		password := getEnv("MONGODB_PASSWORD", "")
		cluster := getEnv("MONGODB_CLUSTER", "atdbcluster0.3ynluj2.mongodb.net")
		dbName = getEnv("MONGODB_DATABASE", "ATDB-cluster")

		fmt.Println("Using MongoDB Atlas")
		serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
		uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=ATDBCluster0",
			username,
			password,
			cluster)
		clientOptions = options.Client().
			ApplyURI(uri).
			SetServerAPIOptions(serverAPIOptions)
	}

	clientOptions.SetServerSelectionTimeout(5 * time.Second)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Println("Failed to create new client")
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		fmt.Println("Failed to connect to client")
		log.Fatal(err)
	}

	// defer func() {
	// 	if err = client.Disconnect(ctx); err != nil {
	// 		panic(err)
	// 	}
	// }()

	Database = client.Database(dbName)

	fmt.Println("Pinging db!")
	if err := client.Ping(ctx, nil); err != nil {
		fmt.Println("Failed to ping database")
		panic(err)
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
