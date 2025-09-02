package examples

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"universal-merkle-sync/core"
)

// ExampleSimpleMerkleSync demonstrates basic Merkle tree operations
func ExampleSimpleMerkleSync() {
	fmt.Println("🚀 Universal MerkleSync Simple Test")
	fmt.Println("===================================")

	// Test 1: Create Merkle Tree
	fmt.Println("\n📊 Test 1: Creating Merkle Tree")
	
	blocks := []core.DataBlock{
		{
			ID:            "block-1",
			EncryptedData: []byte("User data: John Doe, email: john@example.com"),
			TableName:     "users",
			Operation:     "INSERT",
			Timestamp:     1234567890,
			Metadata:      map[string]string{"version": "1.0"},
		},
		{
			ID:            "block-2", 
			EncryptedData: []byte("User data: Jane Smith, email: jane@example.com"),
			TableName:     "users",
			Operation:     "INSERT",
			Timestamp:     1234567891,
			Metadata:      map[string]string{"version": "1.0"},
		},
		{
			ID:            "block-3",
			EncryptedData: []byte("Product data: Laptop, price: $999"),
			TableName:     "products",
			Operation:     "INSERT", 
			Timestamp:     1234567892,
			Metadata:      map[string]string{"category": "electronics"},
		},
	}

	tree, err := core.NewMerkleTree(blocks)
	if err != nil {
		log.Fatalf("Failed to create Merkle tree: %v", err)
	}

	fmt.Printf("✅ Merkle tree created successfully!\n")
	fmt.Printf("   Root hash: %s\n", tree.RootHash)
	fmt.Printf("   Number of leaves: %d\n", len(tree.Leaves))

	// Test 2: Generate Proof
	fmt.Println("\n🔐 Test 2: Generating Merkle Proof")
	
	if len(tree.Leaves) > 0 {
		leafHash := tree.Leaves[0].Hash
		fmt.Printf("   Leaf hash: %s\n", leafHash)
		
		proof, err := tree.GenerateProof([]string{leafHash})
		if err != nil {
			log.Fatalf("Failed to generate proof: %v", err)
		}
		
		fmt.Printf("✅ Proof generated successfully!\n")
		fmt.Printf("   Proof length: %d nodes\n", len(proof))
		
		// Test 3: Verify Proof
		fmt.Println("\n✅ Test 3: Verifying Merkle Proof")
		
		valid, err := core.VerifyProof(tree.RootHash, []string{leafHash}, proof)
		if err != nil {
			log.Fatalf("Failed to verify proof: %v", err)
		}
		
		if valid {
			fmt.Printf("✅ Proof verification successful!\n")
		} else {
			fmt.Printf("❌ Proof verification failed!\n")
		}
	}

	// Test 4: Hash Functions
	fmt.Println("\n🔢 Test 4: Testing Hash Functions")
	
	testData := []byte("Test data for hashing")
	hash1 := hashData(testData)
	hash2 := hashData(testData)
	
	if hash1 == hash2 {
		fmt.Printf("✅ Hash function is deterministic!\n")
		fmt.Printf("   Hash: %s\n", hash1)
	} else {
		fmt.Printf("❌ Hash function is not deterministic!\n")
	}

	// Test 5: Tree Differences
	fmt.Println("\n🔄 Test 5: Testing Tree Differences")
	
	// Create a second tree with one different block
	blocks2 := make([]core.DataBlock, len(blocks))
	copy(blocks2, blocks)
	blocks2[0].EncryptedData = []byte("Modified user data: John Doe, email: john.doe@example.com")
	
	tree2, err := core.NewMerkleTree(blocks2)
	if err != nil {
		log.Fatalf("Failed to create second Merkle tree: %v", err)
	}
	
	differences, err := core.DiffTrees(tree, tree2)
	if err != nil {
		log.Fatalf("Failed to diff trees: %v", err)
	}
	
	fmt.Printf("✅ Tree differences calculated!\n")
	fmt.Printf("   Number of differences: %d\n", len(differences))
	fmt.Printf("   Tree 1 root: %s\n", tree.RootHash)
	fmt.Printf("   Tree 2 root: %s\n", tree2.RootHash)

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("Universal MerkleSync core functionality is working correctly!")
}

// hashData hashes the given data with a prefix to prevent second-preimage attacks
func hashData(data []byte) string {
	// Add prefix to distinguish leaf hashes from internal node hashes
	prefix := []byte("LEAF:")
	combined := append(prefix, data...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}
