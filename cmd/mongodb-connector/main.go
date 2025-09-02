package main

import (
	"context"
	"crypto/rand"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"universal-merkle-sync/connectors/mongodb"
)

func main() {
	connectionString := flag.String("db", "mongodb://localhost:27017", "MongoDB connection string")
	databaseName := flag.String("database", "merklesync", "MongoDB database name")
	grpcServer := flag.String("grpc", "localhost:50051", "gRPC server address")
	flag.Parse()

	// Generate a random encryption key for demo purposes
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create MongoDB connector
	connector, err := mongodb.NewMongoDBConnector(*connectionString, *databaseName, *grpcServer, encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create MongoDB connector: %v", err)
	}
	defer connector.Close()

	// Create demo collection
	err = connector.CreateDemoCollection()
	if err != nil {
		log.Printf("Warning: Failed to create demo collection: %v", err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Start a goroutine to periodically update demo documents
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := connector.UpdateDemoDocument()
				if err != nil {
					log.Printf("Failed to update demo document: %v", err)
				}
			}
		}
	}()

	log.Println("Starting MongoDB connector...")
	
	err = connector.StartChangeStreams(ctx, []string{"users"})
	if err != nil {
		log.Fatalf("MongoDB connector failed: %v", err)
	}
}
