package server

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"universal-merkle-sync/core"
	"universal-merkle-sync/proto"
	"google.golang.org/grpc"
)

// MerkleSyncServer implements the gRPC MerkleSync service
type MerkleSyncServer struct {
	proto.UnimplementedMerkleSyncServer
	blocks      []core.DataBlock
	merkleTree  *core.MerkleTree
	encryptionKey []byte
	mutex       sync.RWMutex
}

// NewMerkleSyncServer creates a new MerkleSync server
func NewMerkleSyncServer(encryptionKey []byte) *MerkleSyncServer {
	return &MerkleSyncServer{
		blocks:        make([]core.DataBlock, 0),
		encryptionKey: encryptionKey,
	}
}

// SubmitBlock handles block submission and Merkle tree updates
func (s *MerkleSyncServer) SubmitBlock(ctx context.Context, req *proto.SubmitBlockRequest) (*proto.SubmitBlockResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Encrypt the data if not already encrypted
	encryptedData := req.Block.EncryptedData
	if len(encryptedData) == 0 {
		// For demo purposes, we'll encrypt the metadata as JSON
		metadataJSON, err := json.Marshal(req.Block.Metadata)
		if err != nil {
			return &proto.SubmitBlockResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to marshal metadata: %v", err),
			}, nil
		}
		
		encryptedData, err = s.encrypt(metadataJSON)
		if err != nil {
			return &proto.SubmitBlockResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to encrypt data: %v", err),
			}, nil
		}
	}

	// Create data block
	block := core.DataBlock{
		ID:            req.Block.Id,
		EncryptedData: encryptedData,
		TableName:     req.Block.TableName,
		Operation:     req.Block.Operation,
		Timestamp:     req.Block.Timestamp,
		Metadata:      req.Block.Metadata,
	}

	// Add to blocks
	s.blocks = append(s.blocks, block)

	// Rebuild Merkle tree
	tree, err := core.NewMerkleTree(s.blocks)
	if err != nil {
		return &proto.SubmitBlockResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to build Merkle tree: %v", err),
		}, nil
	}
	s.merkleTree = tree

	// Calculate leaf hash using core package method for consistency
	leafHash := core.HashData(encryptedData)

	return &proto.SubmitBlockResponse{
		MerkleRoot: s.merkleTree.RootHash,
		LeafHash:   leafHash,
		Success:    true,
	}, nil
}

// GetMerkleRoot returns the current Merkle root
func (s *MerkleSyncServer) GetMerkleRoot(ctx context.Context, req *proto.GetMerkleRootRequest) (*proto.GetMerkleRootResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.merkleTree == nil {
		return &proto.GetMerkleRootResponse{
			MerkleRoot: "",
			BlockCount: 0,
			Timestamp:  time.Now().Unix(),
		}, nil
	}

	blockCount := int64(len(s.blocks))
	if req.TableName != "" {
		// Filter by table name
		count := int64(0)
		for _, block := range s.blocks {
			if block.TableName == req.TableName {
				count++
			}
		}
		blockCount = count
	}

	return &proto.GetMerkleRootResponse{
		MerkleRoot: s.merkleTree.RootHash,
		BlockCount: blockCount,
		Timestamp:  time.Now().Unix(),
	}, nil
}

// GenerateProof generates a Merkle proof for the given leaf hashes
func (s *MerkleSyncServer) GenerateProof(ctx context.Context, req *proto.GenerateProofRequest) (*proto.GenerateProofResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.merkleTree == nil {
		return &proto.GenerateProofResponse{
			Success:      false,
			ErrorMessage: "no Merkle tree available",
		}, nil
	}

	proof, err := s.merkleTree.GenerateProof(req.LeafHashes)
	if err != nil {
		return &proto.GenerateProofResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to generate proof: %v", err),
		}, nil
	}

	// Convert to protobuf format
	proofNodes := make([]*proto.ProofNode, len(proof))
	for i, node := range proof {
		proofNodes[i] = &proto.ProofNode{
			Hash:   node.Hash,
			IsLeft: node.IsLeft,
		}
	}

	return &proto.GenerateProofResponse{
		ProofPath: proofNodes,
		Success:   true,
	}, nil
}

// VerifyProof verifies a Merkle proof
func (s *MerkleSyncServer) VerifyProof(ctx context.Context, req *proto.VerifyProofRequest) (*proto.VerifyProofResponse, error) {
	// Convert protobuf proof to internal format
	proof := make([]core.ProofNode, len(req.ProofPath))
	for i, node := range req.ProofPath {
		proof[i] = core.ProofNode{
			Hash:   node.Hash,
			IsLeft: node.IsLeft,
		}
	}

	valid, err := core.VerifyProof(req.MerkleRoot, req.LeafHashes, proof)
	if err != nil {
		return &proto.VerifyProofResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("verification failed: %v", err),
		}, nil
	}

	return &proto.VerifyProofResponse{
		Valid: valid,
	}, nil
}

// DiffTrees compares two Merkle trees
func (s *MerkleSyncServer) DiffTrees(ctx context.Context, req *proto.DiffTreesRequest) (*proto.DiffTreesResponse, error) {
	// For this implementation, we'll return an error as we need to store multiple trees
	// In a real implementation, you'd have a tree store
	return &proto.DiffTreesResponse{
		Success:      false,
		ErrorMessage: "tree differencing not implemented in this demo",
	}, nil
}

// encrypt encrypts data using AES-GCM
func (s *MerkleSyncServer) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (s *MerkleSyncServer) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// hashData hashes data with prefix (same as in core package)
func (s *MerkleSyncServer) hashData(data []byte) string {
	// Add prefix to distinguish leaf hashes from internal node hashes
	prefix := []byte("LEAF:")
	combined := append(prefix, data...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// StartServer starts the gRPC server
func StartServer(port string, encryptionKey []byte) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	merklesyncServer := NewMerkleSyncServer(encryptionKey)
	proto.RegisterMerkleSyncServer(grpcServer, merklesyncServer)

	log.Printf("Starting MerkleSync gRPC server on port %s", port)
	return grpcServer.Serve(lis)
}
