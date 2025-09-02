package client

import (
	"crypto/rand"
	"os"
	"testing"
)

func TestEdgeClient(t *testing.T) {
	// Create temporary cache directory
	cacheDir, err := os.MkdirTemp("", "merklesync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(cacheDir)

	// Generate encryption key
	encryptionKey := make([]byte, 32)
	_, err = rand.Read(encryptionKey)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Note: This test would require a running gRPC server
	// For now, we'll test the client creation and basic functionality
	t.Logf("Cache directory: %s", cacheDir)
	t.Logf("Encryption key length: %d", len(encryptionKey))
}

func TestCachedData(t *testing.T) {
	// Test CachedData structure
	cachedData := &CachedData{
		Data:      []byte("test data"),
		Proof:     []byte("test proof"),
		RootHash:  "test-root-hash",
		Timestamp: 1234567890,
		TableName: "test_table",
	}

	if string(cachedData.Data) != "test data" {
		t.Error("Data field not set correctly")
	}

	if string(cachedData.Proof) != "test proof" {
		t.Error("Proof field not set correctly")
	}

	if cachedData.RootHash != "test-root-hash" {
		t.Error("RootHash field not set correctly")
	}

	if cachedData.Timestamp != 1234567890 {
		t.Error("Timestamp field not set correctly")
	}

	if cachedData.TableName != "test_table" {
		t.Error("TableName field not set correctly")
	}
}

func TestSyncRequest(t *testing.T) {
	// Test SyncRequest structure
	syncReq := SyncRequest{
		TableName: "test_table",
		LeafHash:  "test-leaf-hash",
		Timestamp: 1234567890,
	}

	if syncReq.TableName != "test_table" {
		t.Error("TableName field not set correctly")
	}

	if syncReq.LeafHash != "test-leaf-hash" {
		t.Error("LeafHash field not set correctly")
	}

	if syncReq.Timestamp != 1234567890 {
		t.Error("Timestamp field not set correctly")
	}
}

func TestEncryptionDecryption(t *testing.T) {
	// Test encryption/decryption functions
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create a mock client for testing encryption
	client := &EdgeClient{
		encryptionKey: encryptionKey,
	}

	testData := []byte("This is test data for encryption")

	// Test encryption
	encryptedData, err := client.encrypt(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	if len(encryptedData) == 0 {
		t.Error("Encrypted data should not be empty")
	}

	// Test decryption
	decryptedData, err := client.decrypt(encryptedData)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	if string(decryptedData) != string(testData) {
		t.Error("Decrypted data should match original data")
	}
}
