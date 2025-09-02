# Universal MerkleSync Technical Specification v1.0

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Components](#core-components)
4. [Data Structures](#data-structures)
5. [API Specifications](#api-specifications)
6. [Implementation Reference](#implementation-reference)
7. [Performance Requirements](#performance-requirements)
8. [Security Model](#security-model)
9. [Deployment Guide](#deployment-guide)

## Overview

Universal MerkleSync is a language-agnostic, database-universal synchronization framework that enables secure, proof-driven, offline-first capabilities for edge computing environments. The system leverages Merkle trees constructed over encrypted data blocks to provide cryptographic integrity guarantees while minimizing bandwidth and storage requirements.

### Key Features
- **Database Agnostic**: Supports PostgreSQL, MongoDB, time-series, graph databases via pluggable connectors
- **Cryptographic Integrity**: Merkle proofs over encrypted ciphertext blocks
- **Offline-First**: O(log N) proof verification without server connectivity
- **Bandwidth Efficient**: Incremental sync with only changed blocks + proofs
- **Language Universal**: gRPC API with generated clients for Go, TypeScript, Python, Rust

## Architecture

[Architecture Diagram - See chart:124]

The system follows a three-tier architecture:

### Tier 1: Database Connectors
Database-specific adapters that monitor change streams and transform native change events into encrypted DataBlock messages.

### Tier 2: MerkleSync Core
Central gRPC server that maintains Merkle trees, generates proofs, and orchestrates synchronization between databases and edge clients.

### Tier 3: Edge Clients  
Lightweight clients that cache encrypted blocks with proofs, verify integrity offline, and perform incremental synchronization.

## Core Components

### 1. Merkle Tree Engine

Based on the high-performance Go implementation from `github.com/txaty/go-merkletree`:

```go
package core

import (
    "crypto/sha256"
    "github.com/txaty/go-merkletree"
    "github.com/BLAKE3-team/BLAKE3/bindings/go"
)

type MerkleConfig struct {
    HashFunc         merkletree.TypeHashFunc
    NumRoutines      int  
    RunInParallel    bool
    SortSiblingPairs bool // OpenZeppelin compatibility
}

func NewMerkleTree(blocks [][]byte, config MerkleConfig) (*merkletree.MerkleTree, error) {
    // Use BLAKE3 for maximum performance
    config.HashFunc = func(data []byte) ([]byte, error) {
        hasher := blake3.New()
        hasher.Write(data)
        return hasher.Sum(nil), nil
    }
    
    config.RunInParallel = true
    config.NumRoutines = 0 // Use all CPU cores
    config.SortSiblingPairs = true
    
    return merkletree.New(&config).Generate(blocks...)
}

func GenerateProof(tree *merkletree.MerkleTree, leafHashes [][]byte) (*merkletree.Proof, error) {
    return tree.GenerateProof(leafHashes)
}

func VerifyProof(proof *merkletree.Proof, root []byte) (bool, error) {
    return merkletree.Verify(proof, root)
}
```

### 2. gRPC Server Implementation

Based on `github.com/johanbrandhorst/grpc-postgres` pattern:

```go
package server

import (
    "context"
    "net"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    pb "github.com/EdgeMerkleDB/universal-merkle-sync/proto"
)

type Server struct {
    pb.UnimplementedMerkleSyncServer
    merkleStore MerkleStore
}

func (s *Server) SubmitBlock(ctx context.Context, req *pb.SubmitBlockRequest) (*pb.SubmitBlockResponse, error) {
    // Validate encrypted block
    if len(req.Block.EncryptedData) == 0 {
        return nil, status.Error(codes.InvalidArgument, "encrypted data required")
    }
    
    // Add to Merkle tree and update root
    blockHash, err := s.merkleStore.AddBlock(req.Block)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to add block: %v", err)
    }
    
    return &pb.SubmitBlockResponse{
        BlockHash: blockHash,
        NewRoot:   s.merkleStore.GetRoot(),
    }, nil
}

func (s *Server) GenerateProof(ctx context.Context, req *pb.GenerateProofRequest) (*pb.GenerateProofResponse, error) {
    proof, err := s.merkleStore.GenerateProof(req.LeafHashes)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "proof generation failed: %v", err)
    }
    
    return &pb.GenerateProofResponse{
        Proof: proof.Serialize(),
        Root:  s.merkleStore.GetRoot(),
    }, nil
}
```

### 3. Database Connectors

#### PostgreSQL Connector (Logical Replication)

```go
package postgresql

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/json"
    "github.com/jackc/pglogrepl"
    "github.com/jackc/pgx/v5"
)

type PostgreSQLConnector struct {
    conn         *pgx.Conn
    grpcClient   pb.MerkleSyncClient
    encryptionKey []byte
    gcm          cipher.AEAD
}

func NewPostgreSQLConnector(connString, grpcAddr string, encKey []byte) (*PostgreSQLConnector, error) {
    // Setup encryption
    block, err := aes.NewCipher(encKey)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    // Connect to PostgreSQL
    conn, err := pgx.Connect(context.Background(), connString)
    if err != nil {
        return nil, err
    }
    
    // Connect to gRPC server
    grpcConn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    return &PostgreSQLConnector{
        conn:         conn,
        grpcClient:   pb.NewMerkleSyncClient(grpcConn),
        encryptionKey: encKey,
        gcm:          gcm,
    }, nil
}

func (p *PostgreSQLConnector) StartReplication(ctx context.Context) error {
    // Create replication slot
    _, err := p.conn.Exec(ctx, "SELECT pg_create_logical_replication_slot('merklesync_slot', 'pgoutput')")
    if err != nil {
        return err
    }
    
    // Start logical replication
    return p.processWALMessages(ctx)
}

func (p *PostgreSQLConnector) processWALMessages(ctx context.Context) error {
    for {
        msg, err := pglogrepl.ReceiveMessage(ctx, p.conn)
        if err != nil {
            return err
        }
        
        switch msg := msg.(type) {
        case *pglogrepl.RelationMessage:
            // Handle schema changes
        case *pglogrepl.InsertMessage:
            p.handleInsert(ctx, msg)
        case *pglogrepl.UpdateMessage:
            p.handleUpdate(ctx, msg)
        case *pglogrepl.DeleteMessage:
            p.handleDelete(ctx, msg)
        }
    }
}

func (p *PostgreSQLConnector) encryptData(data []byte) ([]byte, error) {
    nonce := make([]byte, p.gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }
    
    ciphertext := p.gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}

func (p *PostgreSQLConnector) handleInsert(ctx context.Context, msg *pglogrepl.InsertMessage) error {
    // Serialize row data
    rowData, err := json.Marshal(msg.Tuple)
    if err != nil {
        return err
    }
    
    // Encrypt data
    encryptedData, err := p.encryptData(rowData)
    if err != nil {
        return err
    }
    
    // Submit to MerkleSync
    block := &pb.DataBlock{
        Id:            generateID(),
        EncryptedData: encryptedData,
        TableName:     msg.RelationID,
        Operation:     "INSERT",
        Timestamp:     time.Now().Unix(),
    }
    
    _, err = p.grpcClient.SubmitBlock(ctx, &pb.SubmitBlockRequest{Block: block})
    return err
}
```

#### MongoDB Connector (Change Streams)

```go
package mongodb

import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
)

type MongoDBConnector struct {
    client       *mongo.Client
    db           *mongo.Database
    grpcClient   pb.MerkleSyncClient
    encryptionKey []byte
}

func (m *MongoDBConnector) StartChangeStreams(ctx context.Context, collections []string) error {
    for _, collName := range collections {
        go m.watchCollection(ctx, collName)
    }
    return nil
}

func (m *MongoDBConnector) watchCollection(ctx context.Context, collName string) error {
    collection := m.db.Collection(collName)
    
    // Create change stream
    changeStream, err := collection.Watch(ctx, mongo.Pipeline{})
    if err != nil {
        return err
    }
    defer changeStream.Close(ctx)
    
    for changeStream.Next(ctx) {
        var changeEvent bson.M
        if err := changeStream.Decode(&changeEvent); err != nil {
            continue
        }
        
        m.handleChangeEvent(ctx, changeEvent, collName)
    }
    
    return nil
}

func (m *MongoDBConnector) handleChangeEvent(ctx context.Context, event bson.M, collection string) error {
    // Extract document and operation type
    operationType := event["operationType"].(string)
    fullDocument := event["fullDocument"]
    
    // Serialize document
    docData, err := bson.Marshal(fullDocument)
    if err != nil {
        return err
    }
    
    // Encrypt data
    encryptedData, err := m.encryptData(docData)
    if err != nil {
        return err
    }
    
    // Submit to MerkleSync
    block := &pb.DataBlock{
        Id:            generateID(),
        EncryptedData: encryptedData,
        TableName:     collection,
        Operation:     strings.ToUpper(operationType),
        Timestamp:     time.Now().Unix(),
    }
    
    _, err = m.grpcClient.SubmitBlock(ctx, &pb.SubmitBlockRequest{Block: block})
    return err
}
```

### 4. Edge Client Implementation

```go
package client

import (
    "context"
    "encoding/json"
    "sync"
    "time"
    
    "github.com/dgraph-io/badger/v3"
    pb "github.com/EdgeMerkleDB/universal-merkle-sync/proto"
)

type EdgeClient struct {
    grpcClient    pb.MerkleSyncClient
    cache         *badger.DB
    encryptionKey []byte
    offlineMode   bool
    mu           sync.RWMutex
    
    currentRoot  []byte
    proofCache   map[string]*ProofCacheEntry
}

type ProofCacheEntry struct {
    Proof     []byte
    Timestamp time.Time
    BlockHash []byte
}

func NewEdgeClient(grpcAddr, cacheDir string, encKey []byte) (*EdgeClient, error) {
    // Setup local cache
    opts := badger.DefaultOptions(cacheDir)
    cache, err := badger.Open(opts)
    if err != nil {
        return nil, err
    }
    
    // Connect to gRPC server
    conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    client := &EdgeClient{
        grpcClient:   pb.NewMerkleSyncClient(conn),
        cache:        cache,
        encryptionKey: encKey,
        proofCache:   make(map[string]*ProofCacheEntry),
    }
    
    // Load current root from cache
    client.loadRootFromCache()
    
    return client, nil
}

func (e *EdgeClient) GetData(ctx context.Context, tableName, leafHash string) ([]byte, error) {
    e.mu.RLock()
    defer e.mu.RUnlock()
    
    // Try local cache first
    if data, found := e.getCachedData(leafHash); found {
        // Verify proof if available
        if e.verifyLocalProof(leafHash) {
            return e.decryptData(data)
        }
    }
    
    // If offline, return cached data or error
    if e.offlineMode {
        return nil, ErrDataNotAvailable
    }
    
    // Fetch from server
    return e.fetchFromServer(ctx, tableName, leafHash)
}

func (e *EdgeClient) SyncPending(ctx context.Context) error {
    // Get latest root from server
    rootResp, err := e.grpcClient.GetMerkleRoot(ctx, &pb.GetMerkleRootRequest{})
    if err != nil {
        return err
    }
    
    // Compare with local root
    if bytes.Equal(e.currentRoot, rootResp.Root) {
        return nil // No changes
    }
    
    // Fetch diff and update local cache
    return e.applyDiff(ctx, rootResp.Root)
}

func (e *EdgeClient) verifyLocalProof(leafHash string) bool {
    proofEntry, exists := e.proofCache[leafHash]
    if !exists {
        return false
    }
    
    // Verify proof against current root
    proof, err := DeserializeProof(proofEntry.Proof)
    if err != nil {
        return false
    }
    e
    valid, err := VerifyProof(proof, e.currentRoot)
    return err == nil && valid
}

func (e *EdgeClient) SetOfflineMode(offline bool) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.offlineMode = offline
}
```

## Data Structures

### Protocol Buffer Definitions

```protobuf
syntax = "proto3";

package merklesync.v1;
option go_package = "github.com/EdgeMerkleDB/universal-merkle-sync/proto";

service MerkleSync {
  rpc SubmitBlock(SubmitBlockRequest) returns (SubmitBlockResponse);
  rpc GetMerkleRoot(GetMerkleRootRequest) returns (GetMerkleRootResponse);
  rpc GenerateProof(GenerateProofRequest) returns (GenerateProofResponse);
  rpc VerifyProof(VerifyProofRequest) returns (VerifyProofResponse);
  rpc DiffTrees(DiffTreesRequest) returns (DiffTreesResponse);
  rpc SyncRequest(SyncRequestMessage) returns (stream SyncResponseMessage);
}

message DataBlock {
  string id = 1;
  bytes encrypted_data = 2;
  string table_name = 3;
  string operation = 4; // INSERT, UPDATE, DELETE
  int64 timestamp = 5;
  map<string, string> metadata = 6;
}

message SubmitBlockRequest {
  DataBlock block = 1;
}

message SubmitBlockResponse {
  bytes block_hash = 1;
  bytes new_root = 2;
  bool success = 3;
}

message GetMerkleRootRequest {
  string table_name = 1; // Optional: get root for specific table
}

message GetMerkleRootResponse {
  bytes root = 1;
  int64 timestamp = 2;
  int64 block_count = 3;
}

message GenerateProofRequest {
  repeated bytes leaf_hashes = 1;
}

message GenerateProofResponse {
  bytes proof = 1; // Serialized proof
  bytes root = 2;
  bool valid = 3;
}

message VerifyProofRequest {
  bytes proof = 1;
  bytes root = 2;
  repeated bytes leaf_hashes = 3;
}

message VerifyProofResponse {
  bool valid = 1;
  string error_message = 2;
}

message DiffTreesRequest {
  bytes old_root = 1;
  bytes new_root = 2;
}

message DiffTreesResponse {
  repeated DataBlock changed_blocks = 1;
  repeated bytes deleted_hashes = 2;
  bytes proof_bundle = 3;
}

message SyncRequestMessage {
  bytes current_root = 1;
  repeated string table_filters = 2;
  int64 since_timestamp = 3;
}

message SyncResponseMessage {
  oneof message {
    DiffTreesResponse diff = 1;
    DataBlock block_update = 2;
    bytes new_root = 3;
    bool sync_complete = 4;
  }
}
```

## Performance Requirements

### Merkle Tree Operations

Based on academic benchmarks and open-source implementations:

| Operation | Time Complexity | Performance Target |
|-----------|----------------|-------------------|
| Tree Construction | O(n log n) | 100k leaves in <500ms |
| Proof Generation | O(log n) | Single proof in <1ms |
| Proof Verification | O(log n) | 10k+ verifications/sec |
| Tree Diff | O(n) | 1M leaf diff in <2s |

### Hash Function Performance

Using BLAKE3 for optimal performance:

- **Throughput**: >3 GB/s on modern CPUs
- **Proof Size**: 32 bytes × log₂(n) siblings
- **Memory Usage**: O(n) for tree storage

### Network Efficiency

- **Proof Bundle Size**: ~1KB for trees up to 1M leaves
- **Compression**: gRPC compression reduces payload by 60-80%
- **Batch Operations**: Support up to 10k blocks per batch

## Security Model

### Threat Model

**Assumptions:**
- Network is untrusted and may be adversarial
- Edge devices may be physically compromised
- Database servers are trusted for data integrity
- Cryptographic primitives are secure

**Protections:**
- End-to-end encryption prevents data exposure
- Merkle proofs detect tampering and replay attacks
- Key rotation limits exposure window
- Offline verification reduces attack surface

### Cryptographic Specifications

```go
// Encryption: AES-256-GCM
func NewEncryption(key []byte) (cipher.AEAD, error) {
    block, err := aes.NewCipher(key) // 256-bit key
    if err != nil {
        return nil, err
    }
    return cipher.NewGCM(block)
}

// Hash Function: BLAKE3 (quantum-resistant upgrade available)
func HashFunction(data []byte) []byte {
    return blake3.Sum256(data)
}

// Key Derivation: PBKDF2 with 100,000 iterations
func DeriveKey(password, salt []byte) []byte {
    return pbkdf2.Key(password, salt, 100000, 32, sha256.New)
}
```

### Access Control

- **Client Authentication**: mTLS with client certificates
- **Data Segmentation**: Per-table encryption keys
- **Proof Validation**: Cryptographic verification required
- **Audit Trail**: All operations logged with timestamps

## Deployment Guide

### Docker Compose Setup

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: merklesync
      POSTGRES_USER: postgres 
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    command: postgres -c wal_level=logical

  mongodb:
    image: mongo:7
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: password
    ports:
      - "27017:27017"

  merklesync-server:
    build: ./server
    ports:
      - "50051:50051"
    environment:
      - GRPC_PORT=50051
      - LOG_LEVEL=info
    depends_on:
      - postgres
      - mongodb

  postgresql-connector:
    build: ./connectors/postgresql
    environment:
      - DATABASE_URL=postgres://postgres:password@postgres:5432/merklesync?sslmode=disable
      - GRPC_SERVER=merklesync-server:50051
      - ENCRYPTION_KEY_FILE=/keys/postgres.key
    volumes:
      - ./keys:/keys:ro
    depends_on:
      - merklesync-server
      - postgres

  mongodb-connector:
    build: ./connectors/mongodb  
    environment:
      - MONGODB_URL=mongodb://root:password@mongodb:27017/
      - GRPC_SERVER=merklesync-server:50051
      - ENCRYPTION_KEY_FILE=/keys/mongodb.key
    volumes:
      - ./keys:/keys:ro
    depends_on:
      - merklesync-server
      - mongodb

  edge-client:
    build: ./edge-client
    environment:
      - GRPC_SERVER=merklesync-server:50051
      - CACHE_DIR=/cache
      - ENCRYPTION_KEY_FILE=/keys/client.key
    volumes:
      - ./cache:/cache
      - ./keys:/keys:ro
    depends_on:
      - merklesync-server
```

### Production Configuration

```go
// server/config/config.go
type Config struct {
    GRPCPort        int    `env:"GRPC_PORT" default:"50051"`
    LogLevel        string `env:"LOG_LEVEL" default:"info"`
    MaxMessageSize  int    `env:"MAX_MESSAGE_SIZE" default:"4194304"` // 4MB
    
    // Performance tuning
    MaxConcurrentStreams uint32 `env:"MAX_CONCURRENT_STREAMS" default:"1000"`
    ConnectionTimeout    time.Duration `env:"CONNECTION_TIMEOUT" default:"30s"`
    KeepAliveTime       time.Duration `env:"KEEPALIVE_TIME" default:"30s"`
    
    // Security
    TLSCertFile string `env:"TLS_CERT_FILE"`
    TLSKeyFile  string `env:"TLS_KEY_FILE"`
    ClientCAFile string `env:"CLIENT_CA_FILE"`
    
    // Storage
    BadgerDir      string `env:"BADGER_DIR" default:"/data/badger"`
    MaxTableSize   int64  `env:"MAX_TABLE_SIZE" default:"67108864"` // 64MB
    LevelSizeMultiplier int `env:"LEVEL_SIZE_MULTIPLIER" default:"10"`
}
```

### Monitoring and Observability

```go
// metrics/metrics.go  
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    MerkleTreeSizeGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "merklesync_tree_size_total",
        Help: "Current number of leaves in Merkle tree",
    })
    
    ProofGenerationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "merklesync_proof_generation_duration_seconds",
        Help: "Time taken to generate Merkle proofs",
        Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
    })
    
    SyncRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "merklesync_sync_requests_total",
        Help: "Total number of sync requests by client and status",
    }, []string{"client_id", "status"})
)
```

This specification provides a comprehensive technical foundation for implementing Universal MerkleSync, incorporating proven open-source components and established patterns for production deployment.