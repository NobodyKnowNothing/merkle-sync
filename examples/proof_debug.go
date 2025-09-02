package examples

import (
	"fmt"
	"log"

	"universal-merkle-sync/core"
)

// ExampleProofDebug demonstrates debugging proof generation and verification
func ExampleProofDebug() {
	fmt.Println("🔍 Debugging Proof Generation")
	fmt.Println("=============================")

	// Create test blocks
	blocks := []core.DataBlock{
		{ID: "1", EncryptedData: []byte("data1"), TableName: "test_table"},
		{ID: "2", EncryptedData: []byte("data2"), TableName: "test_table"},
		{ID: "3", EncryptedData: []byte("data3"), TableName: "test_table"},
	}

	// Create tree
	tree, err := core.NewMerkleTree(blocks)
	if err != nil {
		log.Fatalf("Failed to create tree: %v", err)
	}

	fmt.Printf("Tree root: %s\n", tree.RootHash)
	fmt.Printf("Number of leaves: %d\n", len(tree.Leaves))

	// Get all leaf hashes
	leafHashes := make([]string, len(tree.Leaves))
	for i, leaf := range tree.Leaves {
		leafHashes[i] = leaf.Hash
		fmt.Printf("Leaf %d: %s\n", i, leaf.Hash)
	}

	// Generate proof for all leaves
	proof, err := tree.GenerateProof(leafHashes)
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}

	fmt.Printf("Proof length: %d\n", len(proof))
	for i, node := range proof {
		fmt.Printf("Proof node %d: %s (IsLeft: %t)\n", i, node.Hash, node.IsLeft)
	}

	// Try to verify each leaf individually
	fmt.Println("\n🔍 Verifying each leaf individually:")
	for i, leafHash := range leafHashes {
		// Generate proof for just this leaf
		singleProof, err := tree.GenerateProof([]string{leafHash})
		if err != nil {
			log.Fatalf("Failed to generate proof for leaf %d: %v", i, err)
		}
		
		// Verify the proof
		valid, err := core.VerifyProof(tree.RootHash, []string{leafHash}, singleProof)
		if err != nil {
			log.Fatalf("Failed to verify proof for leaf %d: %v", i, err)
		}
		
		fmt.Printf("Leaf %d verification: %t\n", i, valid)
	}

	// Now try to verify all leaves together
	fmt.Println("\n🔍 Verifying all leaves together:")
	valid, err := core.VerifyProof(tree.RootHash, leafHashes, proof)
	if err != nil {
		log.Fatalf("Failed to verify proof for all leaves: %v", err)
	}
	
	fmt.Printf("All leaves verification: %t\n", valid)
}
