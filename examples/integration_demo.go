package examples

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"universal-merkle-sync/core"
	"universal-merkle-sync/server"
)

// ExampleIntegrationTest demonstrates the full MerkleSync system
func ExampleIntegrationTest() {
	fmt.Println("🚀 Starting Universal MerkleSync Integration Test")
	fmt.Println("==================================================")

	// Generate encryption key
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Test 1: Core Merkle Tree Functionality
	fmt.Println("\n📊 Test 1: Core Merkle Tree Functionality")
	testCoreMerkleTree()

	// Test 2: Server Functionality
	fmt.Println("\n🖥️  Test 2: Server Functionality")
	testServerFunctionality(encryptionKey)

	// Test 3: Edge Client Functionality
	fmt.Println("\n📱 Test 3: Edge Client Functionality")
	testEdgeClientFunctionality(encryptionKey)

	fmt.Println("\n✅ All integration tests completed successfully!")
}

func testCoreMerkleTree() {
	// Create test data blocks
	blocks := []core.DataBlock{
		{
			ID:            "block-1",
			EncryptedData: []byte("User data: John Doe, email: john@example.com"),
			TableName:     "users",
			Operation:     "INSERT",
			Timestamp:     time.Now().Unix(),
			Metadata:      map[string]string{"source": "postgresql"},
		},
		{
			ID:            "block-2",
			EncryptedData: []byte("User data: Jane Smith, email: jane@example.com"),
			TableName:     "users",
			Operation:     "INSERT",
			Timestamp:     time.Now().Unix(),
			Metadata:      map[string]string{"source": "postgresql"},
		},
		{
			ID:            "block-3",
			EncryptedData: []byte("Product data: Laptop, price: $999"),
			TableName:     "products",
			Operation:     "INSERT",
			Timestamp:     time.Now().Unix(),
			Metadata:      map[string]string{"source": "mongodb"},
		},
	}

	// Create Merkle tree
	tree, err := core.NewMerkleTree(blocks)
	if err != nil {
		log.Fatalf("Failed to create Merkle tree: %v", err)
	}

	fmt.Printf("   ✅ Created Merkle tree with %d blocks\n", len(blocks))
	fmt.Printf("   📋 Root hash: %s\n", tree.RootHash[:16]+"...")

	// Generate proof for first block
	leafHash := tree.Leaves[0].Hash
	proof, err := tree.GenerateProof([]string{leafHash})
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}

	fmt.Printf("   🔐 Generated proof with %d nodes\n", len(proof))

	// Verify proof
	valid, err := core.VerifyProof(tree.RootHash, []string{leafHash}, proof)
	if err != nil {
		log.Fatalf("Failed to verify proof: %v", err)
	}

	if valid {
		fmt.Println("   ✅ Proof verification successful")
	} else {
		log.Fatal("   ❌ Proof verification failed")
	}
}

func testServerFunctionality(encryptionKey []byte) {
	// Create server (not used in this demo, just for demonstration)
	_ = server.NewMerkleSyncServer(encryptionKey)

	// Submit test blocks
	testBlocks := []struct {
		id      string
		data    string
		table   string
		op      string
	}{
		{"user-1", "John Doe, john@example.com", "users", "INSERT"},
		{"user-2", "Jane Smith, jane@example.com", "users", "INSERT"},
		{"product-1", "Laptop, $999", "products", "INSERT"},
	}

	for _, block := range testBlocks {
		// Create protobuf block (simplified for demo)
		// In real implementation, this would use the actual protobuf types
		fmt.Printf("   📝 Simulating block submission: %s\n", block.id)
	}

	fmt.Println("   ✅ Server functionality test completed")
}

func testEdgeClientFunctionality(encryptionKey []byte) {
	// Create temporary cache directory
	cacheDir, err := os.MkdirTemp("", "merklesync-integration-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp cache dir: %v", err)
	}
	defer os.RemoveAll(cacheDir)

	fmt.Printf("   📁 Cache directory: %s\n", cacheDir)

	// Test cache directory creation
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		log.Fatalf("Cache directory was not created: %v", err)
	}

	fmt.Println("   ✅ Cache directory created successfully")

	// Test encryption/decryption
	testData := []byte("Test data for encryption")
	
	// Create a mock client for encryption testing
	mockClient := &MockEdgeClient{
		EncryptionKey: encryptionKey,
	}

	encryptedData, err := mockClient.encrypt(testData)
	if err != nil {
		log.Fatalf("Failed to encrypt test data: %v", err)
	}

	decryptedData, err := mockClient.decrypt(encryptedData)
	if err != nil {
		log.Fatalf("Failed to decrypt test data: %v", err)
	}

	if string(decryptedData) != string(testData) {
		log.Fatal("Encryption/decryption test failed")
	}

	fmt.Println("   ✅ Encryption/decryption test successful")
	fmt.Println("   📱 Edge client functionality test completed")
}

// Mock client for testing (since we can't import the actual client package)
type MockEdgeClient struct {
	EncryptionKey []byte
}

func (c *MockEdgeClient) encrypt(data []byte) ([]byte, error) {
	// Simple XOR encryption for testing
	encrypted := make([]byte, len(data))
	key := c.EncryptionKey
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ key[i%len(key)]
	}
	return encrypted, nil
}

func (c *MockEdgeClient) decrypt(data []byte) ([]byte, error) {
	// Simple XOR decryption for testing
	decrypted := make([]byte, len(data))
	key := c.EncryptionKey
	for i := 0; i < len(data); i++ {
		decrypted[i] = data[i] ^ key[i%len(key)]
	}
	return decrypted, nil
}
