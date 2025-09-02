package core

import (
	"testing"
)

func TestNewMerkleTree(t *testing.T) {
	// Test empty tree
	tree, err := NewMerkleTree([]DataBlock{})
	if err != nil {
		t.Fatalf("Failed to create empty tree: %v", err)
	}
	if tree.Root != nil {
		t.Error("Empty tree should have nil root")
	}
	if tree.RootHash != "" {
		t.Error("Empty tree should have empty root hash")
	}

	// Test single block
	block := DataBlock{
		ID:            "test-1",
		EncryptedData: []byte("test data"),
		TableName:     "test_table",
		Operation:     "INSERT",
		Timestamp:     1234567890,
		Metadata:      map[string]string{"key": "value"},
	}

	tree, err = NewMerkleTree([]DataBlock{block})
	if err != nil {
		t.Fatalf("Failed to create single block tree: %v", err)
	}
	if tree.Root == nil {
		t.Error("Single block tree should have a root")
	}
	if tree.RootHash == "" {
		t.Error("Single block tree should have a root hash")
	}
	if len(tree.Leaves) != 1 {
		t.Errorf("Expected 1 leaf, got %d", len(tree.Leaves))
	}

	// Test multiple blocks
	blocks := []DataBlock{
		{ID: "1", EncryptedData: []byte("data1"), TableName: "table1"},
		{ID: "2", EncryptedData: []byte("data2"), TableName: "table1"},
		{ID: "3", EncryptedData: []byte("data3"), TableName: "table1"},
		{ID: "4", EncryptedData: []byte("data4"), TableName: "table1"},
	}

	tree, err = NewMerkleTree(blocks)
	if err != nil {
		t.Fatalf("Failed to create multi-block tree: %v", err)
	}
	if tree.Root == nil {
		t.Error("Multi-block tree should have a root")
	}
	if len(tree.Leaves) != 4 {
		t.Errorf("Expected 4 leaves, got %d", len(tree.Leaves))
	}
}

func TestGenerateProof(t *testing.T) {
	blocks := []DataBlock{
		{ID: "1", EncryptedData: []byte("data1"), TableName: "table1"},
		{ID: "2", EncryptedData: []byte("data2"), TableName: "table1"},
		{ID: "3", EncryptedData: []byte("data3"), TableName: "table1"},
		{ID: "4", EncryptedData: []byte("data4"), TableName: "table1"},
	}

	tree, err := NewMerkleTree(blocks)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test proof generation for first leaf
	leafHash := tree.Leaves[0].Hash
	proof, err := tree.GenerateProof([]string{leafHash})
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}
	if len(proof) == 0 {
		t.Error("Proof should not be empty")
	}

	// Test proof generation for multiple leaves
	leafHashes := []string{tree.Leaves[0].Hash, tree.Leaves[2].Hash}
	proof, err = tree.GenerateProof(leafHashes)
	if err != nil {
		t.Fatalf("Failed to generate proof for multiple leaves: %v", err)
	}
	if len(proof) == 0 {
		t.Error("Proof should not be empty")
	}
}

func TestVerifyProof(t *testing.T) {
	blocks := []DataBlock{
		{ID: "1", EncryptedData: []byte("data1"), TableName: "table1"},
		{ID: "2", EncryptedData: []byte("data2"), TableName: "table1"},
		{ID: "3", EncryptedData: []byte("data3"), TableName: "table1"},
		{ID: "4", EncryptedData: []byte("data4"), TableName: "table1"},
	}

	tree, err := NewMerkleTree(blocks)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test valid proof
	leafHash := tree.Leaves[0].Hash
	proof, err := tree.GenerateProof([]string{leafHash})
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	valid, err := VerifyProof(tree.RootHash, []string{leafHash}, proof)
	if err != nil {
		t.Fatalf("Failed to verify proof: %v", err)
	}
	if !valid {
		t.Error("Valid proof should verify successfully")
	}

	// Test invalid proof
	invalidProof := []ProofNode{
		{Hash: "invalid_hash", IsLeft: true},
	}
	valid, err = VerifyProof(tree.RootHash, []string{leafHash}, invalidProof)
	if err != nil {
		t.Fatalf("Failed to verify invalid proof: %v", err)
	}
	if valid {
		t.Error("Invalid proof should not verify successfully")
	}
}

func TestHashData(t *testing.T) {
	data := []byte("test data")
	hash1 := HashData(data)
	hash2 := HashData(data)

	// Same data should produce same hash
	if hash1 != hash2 {
		t.Error("Same data should produce same hash")
	}

	// Different data should produce different hashes
	hash3 := HashData([]byte("different data"))
	if hash1 == hash3 {
		t.Error("Different data should produce different hashes")
	}

	// Hash should not be empty
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}
}

func TestHashConcat(t *testing.T) {
	hash1 := "abc123"
	hash2 := "def456"
	
	result1 := HashConcat(hash1, hash2)
	result2 := HashConcat(hash1, hash2)
	
	// Same inputs should produce same output
	if result1 != result2 {
		t.Error("Same inputs should produce same output")
	}
	
	// Different order should produce different output
	result3 := HashConcat(hash2, hash1)
	if result1 == result3 {
		t.Error("Different order should produce different output")
	}
	
	// Result should not be empty
	if result1 == "" {
		t.Error("Result should not be empty")
	}
}

func TestDiffTrees(t *testing.T) {
	// Create two identical trees
	blocks1 := []DataBlock{
		{ID: "1", EncryptedData: []byte("data1"), TableName: "table1"},
		{ID: "2", EncryptedData: []byte("data2"), TableName: "table1"},
	}
	
	blocks2 := []DataBlock{
		{ID: "1", EncryptedData: []byte("data1"), TableName: "table1"},
		{ID: "2", EncryptedData: []byte("data2"), TableName: "table1"},
	}

	tree1, err := NewMerkleTree(blocks1)
	if err != nil {
		t.Fatalf("Failed to create tree1: %v", err)
	}

	tree2, err := NewMerkleTree(blocks2)
	if err != nil {
		t.Fatalf("Failed to create tree2: %v", err)
	}

	// Identical trees should have no differences
	differences, err := DiffTrees(tree1, tree2)
	if err != nil {
		t.Fatalf("Failed to diff identical trees: %v", err)
	}
	if len(differences) != 0 {
		t.Errorf("Identical trees should have no differences, got %d", len(differences))
	}

	// Create different trees
	blocks3 := []DataBlock{
		{ID: "1", EncryptedData: []byte("data1"), TableName: "table1"},
		{ID: "3", EncryptedData: []byte("data3"), TableName: "table1"},
	}

	tree3, err := NewMerkleTree(blocks3)
	if err != nil {
		t.Fatalf("Failed to create tree3: %v", err)
	}

	// Different trees should have differences
	differences, err = DiffTrees(tree1, tree3)
	if err != nil {
		t.Fatalf("Failed to diff different trees: %v", err)
	}
	if len(differences) == 0 {
		t.Error("Different trees should have differences")
	}
}
