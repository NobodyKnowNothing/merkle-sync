package client

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"universal-merkle-sync/core"
	"universal-merkle-sync/proto"

	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EdgeClient represents an offline-first edge client
type EdgeClient struct {
	grpcClient    proto.MerkleSyncClient
	localDB       *leveldb.DB
	encryptionKey []byte
	cacheDir      string
	mutex         sync.RWMutex
	offlineMode   bool
	pendingSync   []SyncRequest
}

// SyncRequest represents a pending sync operation
type SyncRequest struct {
	TableName string
	LeafHash  string
	Timestamp int64
}

// CachedData represents locally cached data with proof
type CachedData struct {
	Data      []byte `json:"data"`
	Proof     []byte `json:"proof"`
	RootHash  string `json:"root_hash"`
	Timestamp int64  `json:"timestamp"`
	TableName string `json:"table_name"`
}

// NewEdgeClient creates a new edge client
func NewEdgeClient(grpcServerAddr, cacheDir string, encryptionKey []byte) (*EdgeClient, error) {
	// Connect to gRPC server
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	grpcClient := proto.NewMerkleSyncClient(conn)

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Open local database
	dbPath := filepath.Join(cacheDir, "merklesync.db")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open local database: %v", err)
	}

	return &EdgeClient{
		grpcClient:    grpcClient,
		localDB:       db,
		encryptionKey: encryptionKey,
		cacheDir:      cacheDir,
		offlineMode:   false,
		pendingSync:   make([]SyncRequest, 0),
	}, nil
}

// GetData retrieves data with offline-first logic
func (c *EdgeClient) GetData(ctx context.Context, tableName string, leafHash string) (*CachedData, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// First, try to get from local cache
	cachedData, err := c.getFromCache(tableName, leafHash)
	if err == nil && cachedData != nil {
		// Verify the cached data
		valid, err := c.verifyCachedData(cachedData)
		if err == nil && valid {
			log.Printf("Retrieved data from cache for table %s, hash %s", tableName, leafHash)
			return cachedData, nil
		} else {
			log.Printf("Cached data verification failed: %v", err)
		}
	}

	// If not in cache or verification failed, try to fetch from server
	if !c.offlineMode {
		return c.fetchFromServer(ctx, tableName, leafHash)
	}

	// If offline, queue for later sync
	c.queueForSync(tableName, leafHash)
	return nil, fmt.Errorf("data not available offline, queued for sync")
}

// getFromCache retrieves data from local cache
func (c *EdgeClient) getFromCache(tableName, leafHash string) (*CachedData, error) {
	key := fmt.Sprintf("cache:%s:%s", tableName, leafHash)
	data, err := c.localDB.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}

	var cachedData CachedData
	err = json.Unmarshal(data, &cachedData)
	if err != nil {
		return nil, err
	}

	return &cachedData, nil
}

// verifyCachedData verifies the integrity of cached data
func (c *EdgeClient) verifyCachedData(cachedData *CachedData) (bool, error) {
	// Parse the proof
	var proofNodes []core.ProofNode
	err := json.Unmarshal(cachedData.Proof, &proofNodes)
	if err != nil {
		return false, fmt.Errorf("failed to parse proof: %v", err)
	}

	leafHash := core.HashToString(cachedData.Data)

	// Verify the proof
	valid, err := core.VerifyProof(cachedData.RootHash, []string{leafHash}, proofNodes)
	if err != nil {
		return false, fmt.Errorf("proof verification failed: %v", err)
	}

	return valid, nil
}

// fetchFromServer fetches data from the server
func (c *EdgeClient) fetchFromServer(ctx context.Context, tableName, leafHash string) (*CachedData, error) {
	// Get current Merkle root
	rootResp, err := c.grpcClient.GetMerkleRoot(ctx, &proto.GetMerkleRootRequest{
		TableName: tableName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Merkle root: %v", err)
	}

	// Generate proof for the leaf hash
	proofResp, err := c.grpcClient.GenerateProof(ctx, &proto.GenerateProofRequest{
		LeafHashes: []string{leafHash},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %v", err)
	}

	if !proofResp.Success {
		return nil, fmt.Errorf("proof generation failed: %s", proofResp.ErrorMessage)
	}

	// Convert proof to internal format
	proofNodes := make([]core.ProofNode, len(proofResp.ProofPath))
	for i, node := range proofResp.ProofPath {
		hashBytes, err := core.StringToHash(node.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode proof hash: %v", err)
		}
		proofNodes[i] = core.ProofNode{
			Hash:   core.HashToString(hashBytes),
			IsLeft: node.IsLeft,
		}
	}

	// Serialize proof
	proofData, err := json.Marshal(proofNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize proof: %v", err)
	}

	// Create cached data
	cachedData := &CachedData{
		Data:      []byte(leafHash), // In real implementation, this would be the actual data
		Proof:     proofData,
		RootHash:  rootResp.MerkleRoot,
		Timestamp: time.Now().Unix(),
		TableName: tableName,
	}

	// Store in cache
	err = c.storeInCache(tableName, leafHash, cachedData)
	if err != nil {
		log.Printf("Failed to store in cache: %v", err)
	}

	log.Printf("Fetched data from server for table %s, hash %s", tableName, leafHash)
	return cachedData, nil
}

// storeInCache stores data in local cache
func (c *EdgeClient) storeInCache(tableName, leafHash string, data *CachedData) error {
	key := fmt.Sprintf("cache:%s:%s", tableName, leafHash)
	value, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.localDB.Put([]byte(key), value, nil)
}

// queueForSync queues a request for later synchronization
func (c *EdgeClient) queueForSync(tableName, leafHash string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	syncReq := SyncRequest{
		TableName: tableName,
		LeafHash:  leafHash,
		Timestamp: time.Now().Unix(),
	}

	c.pendingSync = append(c.pendingSync, syncReq)
	log.Printf("Queued sync request for table %s, hash %s", tableName, leafHash)
}

// SyncPending synchronizes all pending requests
func (c *EdgeClient) SyncPending(ctx context.Context) error {
	c.mutex.Lock()
	pending := make([]SyncRequest, len(c.pendingSync))
	copy(pending, c.pendingSync)
	c.pendingSync = c.pendingSync[:0] // Clear the slice
	c.mutex.Unlock()

	if len(pending) == 0 {
		return nil
	}

	log.Printf("Syncing %d pending requests", len(pending))

	for _, req := range pending {
		_, err := c.GetData(ctx, req.TableName, req.LeafHash)
		if err != nil {
			log.Printf("Failed to sync request for table %s, hash %s: %v",
				req.TableName, req.LeafHash, err)
			// Re-queue failed requests
			c.queueForSync(req.TableName, req.LeafHash)
		}
	}

	return nil
}

// SetOfflineMode sets the offline mode
func (c *EdgeClient) SetOfflineMode(offline bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.offlineMode = offline
	log.Printf("Offline mode set to: %v", offline)
}

// GetCacheStats returns cache statistics
func (c *EdgeClient) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count cached items
	iter := c.localDB.NewIterator(nil, nil)
	count := 0
	for iter.Next() {
		count++
	}
	iter.Release()

	stats["cached_items"] = count
	stats["pending_sync"] = len(c.pendingSync)
	stats["offline_mode"] = c.offlineMode

	return stats, nil
}

// ClearCache clears the local cache
func (c *EdgeClient) ClearCache() error {
	iter := c.localDB.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		err := c.localDB.Delete(iter.Key(), nil)
		if err != nil {
			return err
		}
	}

	log.Println("Cache cleared")
	return nil
}

// Close closes the edge client
func (c *EdgeClient) Close() error {
	return c.localDB.Close()
}

// encrypt encrypts data using AES-GCM
func (c *EdgeClient) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (c *EdgeClient) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.encryptionKey)
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
