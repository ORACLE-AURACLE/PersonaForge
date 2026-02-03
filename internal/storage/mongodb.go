package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// MongoDatabase wraps the MongoDB client
type MongoDatabase struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoDatabase creates a new MongoDB database connection
func NewMongoDatabase(mongoURI string) (*MongoDatabase, error) {
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI(mongoURI).
		SetServerAPIOptions(serverAPI).
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Majority()).
		SetReadPreference(readpref.Primary())

	// Create a new client and connect to the server
	// In MongoDB v2, Connect takes only opts (context is handled internally)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Send a ping to confirm a successful connection
	ctx := context.Background()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Extract database name from URI or use default
	dbName := "persona-forge-backend"
	// Try to extract from URI if possible
	// For now, use default or parse from URI if needed

	database := client.Database(dbName)

	return &MongoDatabase{
		Client:   client,
		Database: database,
	}, nil
}

// GetSQLDB returns nil for MongoDB
func (d *MongoDatabase) GetSQLDB() *sql.DB {
	return nil
}

// GetMongoClient returns the underlying MongoDB client
func (d *MongoDatabase) GetMongoClient() interface{} {
	return d.Client
}

// RunMigrations sets up MongoDB indexes (equivalent to SQL migrations)
func (d *MongoDatabase) RunMigrations() error {
	ctx := context.TODO()

	// Create indexes for users collection
	usersCollection := d.Database.Collection("users")
	usersIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "google_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
		},
	}
	if _, err := usersCollection.Indexes().CreateMany(ctx, usersIndexes); err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Create indexes for personas collection
	personasCollection := d.Database.Collection("personas")
	personasIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "session_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_default", Value: 1}},
		},
	}
	if _, err := personasCollection.Indexes().CreateMany(ctx, personasIndexes); err != nil {
		return fmt.Errorf("failed to create personas indexes: %w", err)
	}

	// Create indexes for sessions collection
	sessionsCollection := d.Database.Collection("sessions")
	sessionsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
		},
	}
	if _, err := sessionsCollection.Indexes().CreateMany(ctx, sessionsIndexes); err != nil {
		return fmt.Errorf("failed to create sessions indexes: %w", err)
	}

	// Create indexes for messages collection
	messagesCollection := d.Database.Collection("messages")
	messagesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "session_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "persona_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "session_id", Value: 1}, {Key: "persona_id", Value: 1}},
		},
	}
	if _, err := messagesCollection.Indexes().CreateMany(ctx, messagesIndexes); err != nil {
		return fmt.Errorf("failed to create messages indexes: %w", err)
	}

	// Create indexes for token_usage collection
	tokenUsageCollection := d.Database.Collection("token_usage")
	tokenUsageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "session_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}
	if _, err := tokenUsageCollection.Indexes().CreateMany(ctx, tokenUsageIndexes); err != nil {
		return fmt.Errorf("failed to create token_usage indexes: %w", err)
	}

	fmt.Println("MongoDB indexes created successfully")
	return nil
}

// Close closes the MongoDB connection
func (d *MongoDatabase) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.Client.Disconnect(ctx)
}
