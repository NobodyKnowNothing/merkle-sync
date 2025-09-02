package main

import (
	"context"
	"crypto/rand"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"universal-merkle-sync/connectors/postgresql"
)

func main() {
	connectionString := flag.String("db", "postgres://user:password@localhost:5432/merklesync?sslmode=disable", "PostgreSQL connection string")
	grpcServer := flag.String("grpc", "localhost:50051", "gRPC server address")
	flag.Parse()

	// Generate a random encryption key for demo purposes
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create PostgreSQL connector
	connector, err := postgresql.NewPostgreSQLConnector(*connectionString, *grpcServer, encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL connector: %v", err)
	}

	// Create demo table
	err = connector.CreateDemoTable()
	if err != nil {
		log.Printf("Warning: Failed to create demo table: %v", err)
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

	log.Println("Starting PostgreSQL connector...")
	
	err = connector.StartReplication(ctx)
	if err != nil {
		log.Fatalf("PostgreSQL connector failed: %v", err)
	}
}
