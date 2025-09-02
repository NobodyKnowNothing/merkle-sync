package main

import (
	"context"
	"crypto/rand"
	"flag"
	"log"
	"time"

	"universal-merkle-sync/edge-client"
)

func main() {
	grpcServer := flag.String("grpc", "localhost:50051", "gRPC server address")
	cacheDir := flag.String("cache", "./cache", "Cache directory")
	tableName := flag.String("table", "users", "Table name to query")
	leafHash := flag.String("hash", "demo-hash", "Leaf hash to query")
	flag.Parse()

	// Generate a random encryption key for demo purposes
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create edge client
	edgeClient, err := client.NewEdgeClient(*grpcServer, *cacheDir, encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create edge client: %v", err)
	}
	defer edgeClient.Close()

	ctx := context.Background()

	// Test online mode
	log.Println("Testing online mode...")
	data, err := edgeClient.GetData(ctx, *tableName, *leafHash)
	if err != nil {
		log.Printf("Failed to get data online: %v", err)
	} else {
		log.Printf("Successfully retrieved data: %+v", data)
	}

	// Test offline mode
	log.Println("Testing offline mode...")
	edgeClient.SetOfflineMode(true)
	
	data, err = edgeClient.GetData(ctx, *tableName, *leafHash)
	if err != nil {
		log.Printf("Expected error in offline mode: %v", err)
	} else {
		log.Printf("Retrieved data from cache: %+v", data)
	}

	// Test sync pending
	log.Println("Testing sync pending...")
	edgeClient.SetOfflineMode(false)
	err = edgeClient.SyncPending(ctx)
	if err != nil {
		log.Printf("Failed to sync pending: %v", err)
	} else {
		log.Println("Successfully synced pending requests")
	}

	// Get cache stats
	stats, err := edgeClient.GetCacheStats()
	if err != nil {
		log.Printf("Failed to get cache stats: %v", err)
	} else {
		log.Printf("Cache stats: %+v", stats)
	}

	// Keep running for demo
	log.Println("Edge client demo running... Press Ctrl+C to exit")
	for {
		time.Sleep(10 * time.Second)
		
		// Try to get data periodically
		data, err := edgeClient.GetData(ctx, *tableName, *leafHash)
		if err != nil {
			log.Printf("Failed to get data: %v", err)
		} else {
			log.Printf("Retrieved data: %+v", data)
		}
	}
}
