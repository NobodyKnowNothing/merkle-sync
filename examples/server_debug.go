package examples

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"

	"universal-merkle-sync/server"
	"universal-merkle-sync/proto"
)

// ExampleServerDebug demonstrates debugging server proof verification
func ExampleServerDebug() {
	fmt.Println("🖥️ Debugging Server Proof Verification")
	fmt.Println("======================================")

	// Generate encryption key
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create server
	merklesyncServer := server.NewMerkleSyncServer(encryptionKey)

	// Submit blocks
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
		resp, err := merklesyncServer.SubmitBlock(context.Background(), req)
		if err != nil {
			log.Fatalf("Failed to submit block %s: %v", block.Id, err)
		}
		if !resp.Success {
			log.Fatalf("Submit block %s failed: %s", block.Id, resp.ErrorMessage)
		}
		leafHashes = append(leafHashes, resp.LeafHash)
		fmt.Printf("Block %s: Leaf hash = %s\n", block.Id, resp.LeafHash)
	}

	// Get Merkle root
	rootReq := &proto.GetMerkleRootRequest{TableName: "test_table"}
	rootResp, err := merklesyncServer.GetMerkleRoot(context.Background(), rootReq)
	if err != nil {
		log.Fatalf("Failed to get Merkle root: %v", err)
	}

	fmt.Printf("Merkle root: %s\n", rootResp.MerkleRoot)
	fmt.Printf("Block count: %d\n", rootResp.BlockCount)

	// Generate proof for all leaves
	proofReq := &proto.GenerateProofRequest{
		MerkleRoot:  rootResp.MerkleRoot,
		LeafHashes: leafHashes,
	}

	proofResp, err := merklesyncServer.GenerateProof(context.Background(), proofReq)
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}

	if !proofResp.Success {
		log.Fatalf("Generate proof failed: %s", proofResp.ErrorMessage)
	}

	fmt.Printf("Proof length: %d\n", len(proofResp.ProofPath))
	for i, node := range proofResp.ProofPath {
		fmt.Printf("Proof node %d: %s (IsLeft: %t)\n", i, node.Hash, node.IsLeft)
	}

	// Verify proof
	verifyReq := &proto.VerifyProofRequest{
		MerkleRoot:  rootResp.MerkleRoot,
		LeafHashes: leafHashes,
		ProofPath:   proofResp.ProofPath,
	}

	verifyResp, err := merklesyncServer.VerifyProof(context.Background(), verifyReq)
	if err != nil {
		log.Fatalf("Failed to verify proof: %v", err)
	}

	fmt.Printf("Proof verification result: %t\n", verifyResp.Valid)
	if !verifyResp.Valid {
		fmt.Printf("Error message: %s\n", verifyResp.ErrorMessage)
	}
}
