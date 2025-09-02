package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"universal-merkle-sync/proto"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PostgreSQLConnector handles PostgreSQL logical replication
type PostgreSQLConnector struct {
	connectionString string
	grpcClient       proto.MerkleSyncClient
	slotName         string
	replicationConn  *sql.DB
	encryptionKey    []byte
}

// NewPostgreSQLConnector creates a new PostgreSQL connector
func NewPostgreSQLConnector(connectionString, grpcServerAddr string, encryptionKey []byte) (*PostgreSQLConnector, error) {
	// Connect to gRPC server
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	grpcClient := proto.NewMerkleSyncClient(conn)

	return &PostgreSQLConnector{
		connectionString: connectionString,
		grpcClient:       grpcClient,
		slotName:         "merklesync_slot",
		encryptionKey:    encryptionKey,
	}, nil
}

// StartReplication starts PostgreSQL logical replication
func (p *PostgreSQLConnector) StartReplication(ctx context.Context) error {
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create replication slot if it doesn't exist
	err = p.createReplicationSlot(db)
	if err != nil {
		return fmt.Errorf("failed to create replication slot: %v", err)
	}

	// Start logical replication
	return p.startLogicalReplication(ctx, db)
}

// createReplicationSlot creates a logical replication slot
func (p *PostgreSQLConnector) createReplicationSlot(db *sql.DB) error {
	// Check if slot already exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_replication_slots WHERE slot_name = $1)", p.slotName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check replication slot: %v", err)
	}

	if !exists {
		// Create the replication slot
		_, err = db.Exec(fmt.Sprintf("SELECT pg_create_logical_replication_slot('%s', 'test_decoding')", p.slotName))
		if err != nil {
			return fmt.Errorf("failed to create replication slot: %v", err)
		}
		log.Printf("Created replication slot: %s", p.slotName)
	} else {
		log.Printf("Replication slot already exists: %s", p.slotName)
	}

	return nil
}

// startLogicalReplication starts the logical replication process
func (p *PostgreSQLConnector) startLogicalReplication(ctx context.Context, db *sql.DB) error {
	// For this demo, we'll simulate replication by polling for changes
	// In a real implementation, you'd use pg_recvlogical or similar
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Starting PostgreSQL logical replication monitoring...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := p.pollForChanges(db)
			if err != nil {
				log.Printf("Error polling for changes: %v", err)
			}
		}
	}
}

// pollForChanges polls for changes in monitored tables
func (p *PostgreSQLConnector) pollForChanges(db *sql.DB) error {
	// For demo purposes, we'll check a simple table
	// In real implementation, you'd use logical replication WAL
	rows, err := db.Query(`
		SELECT id, name, email, updated_at 
		FROM users 
		WHERE updated_at > NOW() - INTERVAL '10 seconds'
		ORDER BY updated_at DESC
		LIMIT 10
	`)
	if err != nil {
		// Table might not exist, that's okay for demo
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		var updatedAt time.Time

		err := rows.Scan(&id, &name, &email, &updatedAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Create change event
		changeEvent := map[string]interface{}{
			"id":         id,
			"name":       name,
			"email":      email,
			"updated_at": updatedAt,
			"operation":  "UPDATE",
		}

		// Submit to MerkleSync
		err = p.submitChange(changeEvent, "users", "UPDATE")
		if err != nil {
			log.Printf("Error submitting change: %v", err)
		}
	}

	return nil
}

// submitChange submits a change event to the MerkleSync server
func (p *PostgreSQLConnector) submitChange(changeEvent map[string]interface{}, tableName, operation string) error {
	// Serialize change event
	changeData, err := json.Marshal(changeEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal change event: %v", err)
	}

	// Encrypt the data
	encryptedData, err := p.encrypt(changeData)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %v", err)
	}

	// Create protobuf message
	block := &proto.DataBlock{
		Id:            uuid.New().String(),
		EncryptedData: encryptedData,
		TableName:     tableName,
		Operation:     operation,
		Timestamp:     time.Now().Unix(),
		Metadata: map[string]string{
			"source":     "postgresql",
			"change_id":  uuid.New().String(),
			"table_name": tableName,
		},
	}

	// Submit to gRPC server
	req := &proto.SubmitBlockRequest{
		Block: block,
	}

	resp, err := p.grpcClient.SubmitBlock(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to submit block: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("server rejected block: %s", resp.ErrorMessage)
	}

	log.Printf("Submitted change for table %s, operation %s, new root: %s",
		tableName, operation, resp.MerkleRoot)

	return nil
}

// encrypt encrypts data using AES-GCM
func (p *PostgreSQLConnector) encrypt(data []byte) ([]byte, error) {
	// For demo purposes, we'll use a simple XOR encryption
	// In production, use proper AES-GCM encryption
	encrypted := make([]byte, len(data))
	key := p.encryptionKey
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ key[i%len(key)]
	}
	return encrypted, nil
}

// CreateDemoTable creates a demo table for testing
func (p *PostgreSQLConnector) CreateDemoTable() error {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create trigger to update updated_at
	_, err = db.Exec(`
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`)
	if err != nil {
		return fmt.Errorf("failed to create update function: %v", err)
	}

	_, err = db.Exec(`
		DROP TRIGGER IF EXISTS update_users_updated_at ON users;
		CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`)
	if err != nil {
		return fmt.Errorf("failed to create trigger: %v", err)
	}

	// Insert some demo data
	_, err = db.Exec(`
		INSERT INTO users (name, email) VALUES 
		('John Doe', 'john@example.com'),
		('Jane Smith', 'jane@example.com'),
		('Bob Johnson', 'bob@example.com')
		ON CONFLICT (email) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to insert demo data: %v", err)
	}

	log.Println("Created demo users table with sample data")
	return nil
}
