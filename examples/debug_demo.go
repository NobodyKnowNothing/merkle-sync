package examples

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sort"

	"universal-merkle-sync/core"
)

// ExampleDebugMerkleProof demonstrates debugging Merkle proof verification
func ExampleDebugMerkleProof() {
	fmt.Println("🔍 Debugging Merkle Proof Verification")
	fmt.Println("=====================================")

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

	// Verify proof
	valid, err := core.VerifyProof(tree.RootHash, leafHashes, proof)
	if err != nil {
		log.Fatalf("Failed to verify proof: %v", err)
	}

	fmt.Printf("Proof verification result: %t\n", valid)

	if !valid {
		fmt.Println("❌ Proof verification failed!")
		
		// Try to debug by reconstructing the root
		computedRoot := reconstructRoot(leafHashes, proof)
		fmt.Printf("Computed root: %s\n", computedRoot)
		fmt.Printf("Expected root: %s\n", tree.RootHash)
		fmt.Printf("Roots match: %t\n", computedRoot == tree.RootHash)
	} else {
		fmt.Println("✅ Proof verification successful!")
	}
}

// Copy the reconstructRoot function for debugging
func reconstructRoot(leafHashes []string, proofPath []core.ProofNode) string {
	if len(leafHashes) == 0 {
		return ""
	}

	// For multiple leaf hashes, we need to reconstruct the entire tree
	if len(leafHashes) > 1 {
		return reconstructRootMultiple(leafHashes, proofPath)
	}

	// For single leaf hash, reconstruct the path to root
	currentHash := leafHashes[0]
	
	for _, proofNode := range proofPath {
		if proofNode.IsLeft {
			// Proof node is on the left, current hash is on the right
			currentHash = hashConcat(proofNode.Hash, currentHash)
		} else {
			// Proof node is on the right, current hash is on the left
			currentHash = hashConcat(currentHash, proofNode.Hash)
		}
	}
	
	return currentHash
}

func reconstructRootMultiple(leafHashes []string, proofPath []core.ProofNode) string {
	// Start with leaf hashes
	currentLevel := make([]string, len(leafHashes))
	copy(currentLevel, leafHashes)
	
	// Sort for consistent ordering
	sort.Strings(currentLevel)
	
	// Create a map to track which proof nodes we've used
	usedProofNodes := make(map[int]bool)
	
	// Build tree level by level
	for len(currentLevel) > 1 {
		nextLevel := make([]string, 0)
		
		// Process pairs
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				// We have a pair, combine them
				combined := hashConcat(currentLevel[i], currentLevel[i+1])
				nextLevel = append(nextLevel, combined)
			} else {
				// Odd number of nodes, we need a proof node
				// Find an unused proof node
				var proofNode *core.ProofNode
				var proofIndex int
				for j, node := range proofPath {
					if !usedProofNodes[j] {
						proofNode = &node
						proofIndex = j
						break
					}
				}
				
				if proofNode != nil {
					usedProofNodes[proofIndex] = true
					
					// Combine with proof node
					var combined string
					if proofNode.IsLeft {
						combined = hashConcat(proofNode.Hash, currentLevel[i])
					} else {
						combined = hashConcat(currentLevel[i], proofNode.Hash)
					}
					nextLevel = append(nextLevel, combined)
				} else {
					// No more proof nodes, keep the last one
					nextLevel = append(nextLevel, currentLevel[i])
				}
			}
		}
		
		currentLevel = nextLevel
	}
	
	if len(currentLevel) == 1 {
		return currentLevel[0]
	}
	
	return ""
}

func hashConcat(hash1, hash2 string) string {
	// Add prefix to distinguish internal node hashes
	prefix := []byte("INTERNAL:")
	combined := append(prefix, []byte(hash1)...)
	combined = append(combined, []byte(hash2)...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}
