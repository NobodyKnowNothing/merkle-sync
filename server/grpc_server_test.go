package server

import (
	"context"
	"crypto/rand"
	"testing"

	"universal-merkle-sync/proto"
)

func TestMerkleSyncServer(t *testing.T) {
	// Generate encryption key
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create server
	server := NewMerkleSyncServer(encryptionKey)

	// Test SubmitBlock
	block := &proto.DataBlock{
		Id:            "test-block-1",
		EncryptedData: []byte("encrypted test data"),
		TableName:     "test_table",
		Operation:     "INSERT",
		Timestamp:     1234567890,
		Metadata: map[string]string{
			"source": "test",
		},
	}

	req := &proto.SubmitBlockRequest{
		Block: block,
	}

	resp, err := server.SubmitBlock(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to submit block: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Submit block failed: %s", resp.ErrorMessage)
	}

	if resp.MerkleRoot == "" {
		t.Error("Merkle root should not be empty")
	}

	if resp.LeafHash == "" {
		t.Error("Leaf hash should not be empty")
	}

	// Test GetMerkleRoot
	rootReq := &proto.GetMerkleRootRequest{
		TableName: "test_table",
	}

	rootResp, err := server.GetMerkleRoot(context.Background(), rootReq)
	if err != nil {
		t.Fatalf("Failed to get Merkle root: %v", err)
	}

	if rootResp.MerkleRoot == "" {
		t.Error("Merkle root should not be empty")
	}

	if rootResp.BlockCount != 1 {
		t.Errorf("Expected 1 block, got %d", rootResp.BlockCount)
	}

	// Test GenerateProof
	proofReq := &proto.GenerateProofRequest{
		MerkleRoot:  rootResp.MerkleRoot,
		LeafHashes: []string{resp.LeafHash},
	}

	proofResp, err := server.GenerateProof(context.Background(), proofReq)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	if !proofResp.Success {
		t.Fatalf("Generate proof failed: %s", proofResp.ErrorMessage)
	}

	// Test VerifyProof
	verifyReq := &proto.VerifyProofRequest{
		MerkleRoot:  rootResp.MerkleRoot,
		LeafHashes: []string{resp.LeafHash},
		ProofPath:   proofResp.ProofPath,
	}

	verifyResp, err := server.VerifyProof(context.Background(), verifyReq)
	if err != nil {
		t.Fatalf("Failed to verify proof: %v", err)
	}

	if !verifyResp.Valid {
		t.Error("Valid proof should verify successfully")
	}
}

func TestMerkleSyncServerMultipleBlocks(t *testing.T) {
	// Generate encryption key
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create server
	server := NewMerkleSyncServer(encryptionKey)

	// Submit multiple blocks
	blocks := []*proto.DataBlock{
		{
			Id:            "block-1",
			EncryptedData: []byte("data1"),
			TableName:     "test_table",
			Operation:     "INSERT",
			Timestamp:     1234567890,
		},
		{
			Id:            "block-2",
			EncryptedData: []byte("data2"),
			TableName:     "test_table",
			Operation:     "INSERT",
			Timestamp:     1234567891,
		},
		{
			Id:            "block-3",
			EncryptedData: []byte("data3"),
			TableName:     "test_table",
			Operation:     "UPDATE",
			Timestamp:     1234567892,
		},
	}

	var leafHashes []string
	for _, block := range blocks {
		req := &proto.SubmitBlockRequest{Block: block}
		resp, err := server.SubmitBlock(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to submit block %s: %v", block.Id, err)
		}
		if !resp.Success {
			t.Fatalf("Submit block %s failed: %s", block.Id, resp.ErrorMessage)
		}
		leafHashes = append(leafHashes, resp.LeafHash)
		t.Logf("Block %s: Leaf hash = %s", block.Id, resp.LeafHash)
	}

	// Get Merkle root
	rootReq := &proto.GetMerkleRootRequest{TableName: "test_table"}
	rootResp, err := server.GetMerkleRoot(context.Background(), rootReq)
	if err != nil {
		t.Fatalf("Failed to get Merkle root: %v", err)
	}

	t.Logf("Merkle root: %s", rootResp.MerkleRoot)
	t.Logf("Block count: %d", rootResp.BlockCount)

	if rootResp.BlockCount != 3 {
		t.Errorf("Expected 3 blocks, got %d", rootResp.BlockCount)
	}

	// Test individual leaf proofs instead of multiple leaf proof
	for i, leafHash := range leafHashes {
		// Generate proof for individual leaf
		proofReq := &proto.GenerateProofRequest{
			MerkleRoot:  rootResp.MerkleRoot,
			LeafHashes: []string{leafHash},
		}

		proofResp, err := server.GenerateProof(context.Background(), proofReq)
		if err != nil {
			t.Fatalf("Failed to generate proof for leaf %d: %v", i, err)
		}

		if !proofResp.Success {
			t.Fatalf("Generate proof failed for leaf %d: %s", i, proofResp.ErrorMessage)
		}

		t.Logf("Proof for leaf %d length: %d", i, len(proofResp.ProofPath))

		// Verify individual leaf proof
		verifyReq := &proto.VerifyProofRequest{
			MerkleRoot:  rootResp.MerkleRoot,
			LeafHashes: []string{leafHash},
			ProofPath:   proofResp.ProofPath,
		}

		verifyResp, err := server.VerifyProof(context.Background(), verifyReq)
		if err != nil {
			t.Fatalf("Failed to verify proof for leaf %d: %v", i, err)
		}

		t.Logf("Verification result for leaf %d: %t", i, verifyResp.Valid)
		if verifyResp.ErrorMessage != "" {
			t.Logf("Verification error for leaf %d: %s", i, verifyResp.ErrorMessage)
		}

		if !verifyResp.Valid {
			t.Errorf("Valid proof for leaf %d should verify successfully", i)
		}
	}
}
