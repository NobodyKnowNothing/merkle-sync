package main

import (
	"crypto/rand"
	"log"
	"universal-merkle-sync/server"
)

func main() {
	// Generate a secure encryption key for the server session.
	// In a production environment, this should be managed securely (e.g., via secrets management).
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Start the gRPC server on the default port.
	if err := server.StartServer("50051", encryptionKey); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
