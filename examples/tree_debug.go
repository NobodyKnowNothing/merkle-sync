package examples

import (
	"fmt"
	"log"

	"universal-merkle-sync/core"
)

// ExampleTreeDebug demonstrates debugging Merkle tree structure
func ExampleTreeDebug() {
	fmt.Println("🌳 Debugging Merkle Tree Structure")
	fmt.Println("==================================")

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

	// Print tree structure
	printTree(tree.Root, 0)

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
}

func printTree(node *core.MerkleNode, level int) {
	if node == nil {
		return
	}
	
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}
	
	if node.IsLeaf {
		fmt.Printf("%sLeaf: %s (ID: %s)\n", indent, node.Hash, node.BlockID)
	} else {
		fmt.Printf("%sInternal: %s\n", indent, node.Hash)
	}
	
	printTree(node.Left, level+1)
	printTree(node.Right, level+1)
}
