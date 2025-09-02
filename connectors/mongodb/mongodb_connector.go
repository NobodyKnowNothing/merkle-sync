package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"universal-merkle-sync/proto"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MongoDBConnector handles MongoDB change streams
type MongoDBConnector struct {
	client        *mongo.Client
	database      *mongo.Database
	grpcClient    proto.MerkleSyncClient
	encryptionKey []byte
}

// NewMongoDBConnector creates a new MongoDB connector
func NewMongoDBConnector(connectionString, databaseName, grpcServerAddr string, encryptionKey []byte) (*MongoDBConnector, error) {
	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Test the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	database := client.Database(databaseName)

	// Connect to gRPC server
	grpcConn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	grpcClient := proto.NewMerkleSyncClient(grpcConn)

	return &MongoDBConnector{
		client:        client,
		database:      database,
		grpcClient:    grpcClient,
		encryptionKey: encryptionKey,
	}, nil
}

// StartChangeStreams starts monitoring MongoDB change streams
func (m *MongoDBConnector) StartChangeStreams(ctx context.Context, collections []string) error {
	log.Printf("Starting MongoDB change streams for collections: %v", collections)

	// Create change stream for all collections
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"operationType", bson.D{
				{"$in", []string{"insert", "update", "delete", "replace"}},
			}},
		}}},
	}

	// Start change stream on the database
	changeStream, err := m.database.Watch(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to create change stream: %v", err)
	}
	defer changeStream.Close(ctx)

	log.Println("MongoDB change stream started successfully")

	// Process change events
	for changeStream.Next(ctx) {
		var changeEvent bson.M
		if err := changeStream.Decode(&changeEvent); err != nil {
			log.Printf("Error decoding change event: %v", err)
			continue
		}

		err := m.processChangeEvent(changeEvent)
		if err != nil {
			log.Printf("Error processing change event: %v", err)
		}
	}

	if err := changeStream.Err(); err != nil {
		return fmt.Errorf("change stream error: %v", err)
	}

	return nil
}

// processChangeEvent processes a single change event
func (m *MongoDBConnector) processChangeEvent(changeEvent bson.M) error {
	// Extract change event details
	operationType, ok := changeEvent["operationType"].(string)
	if !ok {
		return fmt.Errorf("invalid operation type")
	}

	collection, ok := changeEvent["ns"].(bson.M)
	if !ok {
		return fmt.Errorf("invalid namespace")
	}

	collectionName, ok := collection["coll"].(string)
	if !ok {
		return fmt.Errorf("invalid collection name")
	}

	// Extract document data based on operation type
	var documentData bson.M
	var documentID string

	switch operationType {
	case "insert", "replace":
		if fullDocument, exists := changeEvent["fullDocument"]; exists {
			documentData = fullDocument.(bson.M)
		}
	case "update":
		if fullDocument, exists := changeEvent["fullDocument"]; exists {
			documentData = fullDocument.(bson.M)
		} else if documentKey, exists := changeEvent["documentKey"]; exists {
			documentData = documentKey.(bson.M)
		}
	case "delete":
		if documentKey, exists := changeEvent["documentKey"]; exists {
			documentData = documentKey.(bson.M)
		}
	}

	// Extract document ID
	if id, exists := documentData["_id"]; exists {
		if objectID, ok := id.(primitive.ObjectID); ok {
			documentID = objectID.Hex()
		} else {
			documentID = fmt.Sprintf("%v", id)
		}
	}

	// Create change event for MerkleSync
	changeData := map[string]interface{}{
		"operation_type": operationType,
		"collection":     collectionName,
		"document_id":    documentID,
		"document":       documentData,
		"timestamp":      time.Now().Unix(),
	}

	// Submit to MerkleSync
	err := m.submitChange(changeData, collectionName, operationType)
	if err != nil {
		return fmt.Errorf("failed to submit change: %v", err)
	}

	log.Printf("Processed %s operation on collection %s, document %s", 
		operationType, collectionName, documentID)

	return nil
}

// submitChange submits a change event to the MerkleSync server
func (m *MongoDBConnector) submitChange(changeEvent map[string]interface{}, collectionName, operation string) error {
	// Serialize change event
	changeData, err := json.Marshal(changeEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal change event: %v", err)
	}

	// Encrypt the data
	encryptedData, err := m.encrypt(changeData)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %v", err)
	}

	// Create protobuf message
	block := &proto.DataBlock{
		Id:            uuid.New().String(),
		EncryptedData: encryptedData,
		TableName:     collectionName,
		Operation:     operation,
		Timestamp:     time.Now().Unix(),
		Metadata: map[string]string{
			"source":     "mongodb",
			"change_id":  uuid.New().String(),
			"collection": collectionName,
		},
	}

	// Submit to gRPC server
	req := &proto.SubmitBlockRequest{
		Block: block,
	}

	resp, err := m.grpcClient.SubmitBlock(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to submit block: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("server rejected block: %s", resp.ErrorMessage)
	}

	log.Printf("Submitted change for collection %s, operation %s, new root: %s", 
		collectionName, operation, resp.MerkleRoot)

	return nil
}

// encrypt encrypts data using simple XOR (for demo purposes)
func (m *MongoDBConnector) encrypt(data []byte) ([]byte, error) {
	// For demo purposes, we'll use a simple XOR encryption
	// In production, use proper AES-GCM encryption
	encrypted := make([]byte, len(data))
	key := m.encryptionKey
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ key[i%len(key)]
	}
	return encrypted, nil
}

// CreateDemoCollection creates a demo collection for testing
func (m *MongoDBConnector) CreateDemoCollection() error {
	collection := m.database.Collection("users")

	// Insert some demo documents
	documents := []interface{}{
		bson.M{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
			"city":  "New York",
		},
		bson.M{
			"name":  "Jane Smith",
			"email": "jane@example.com",
			"age":   25,
			"city":  "San Francisco",
		},
		bson.M{
			"name":  "Bob Johnson",
			"email": "bob@example.com",
			"age":   35,
			"city":  "Chicago",
		},
	}

	_, err := collection.InsertMany(context.Background(), documents)
	if err != nil {
		return fmt.Errorf("failed to insert demo documents: %v", err)
	}

	log.Println("Created demo users collection with sample documents")
	return nil
}

// UpdateDemoDocument updates a demo document to trigger change stream
func (m *MongoDBConnector) UpdateDemoDocument() error {
	collection := m.database.Collection("users")

	// Update a random document
	filter := bson.M{"name": "John Doe"}
	update := bson.M{
		"$set": bson.M{
			"last_updated": time.Now(),
			"version":      time.Now().Unix(),
		},
	}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %v", err)
	}

	log.Printf("Updated %d document(s) in users collection", result.ModifiedCount)
	return nil
}

// Close closes the MongoDB connection
func (m *MongoDBConnector) Close() error {
	return m.client.Disconnect(context.Background())
}
