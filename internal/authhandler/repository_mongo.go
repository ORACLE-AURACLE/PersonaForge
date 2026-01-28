package authhandler

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/PersonaForge/backend/internal/storage"
)

// MongoRepository handles authentication data access for MongoDB
type MongoRepository struct {
	db *storage.MongoDatabase
}

// NewMongoRepository creates a new auth repository for MongoDB
func NewMongoRepository(db *storage.MongoDatabase) *MongoRepository {
	return &MongoRepository{db: db}
}

// getNextID gets the next sequential ID for a collection
func (r *MongoRepository) getNextID(ctx context.Context, collectionName string) (int, error) {
	counterCollection := r.db.Database.Collection("counters")

	filter := bson.M{"_id": collectionName}
	update := bson.M{
		"$inc":         bson.M{"seq": 1},
		"$setOnInsert": bson.M{"seq": 1},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter struct {
		ID  string `bson:"_id"`
		Seq int    `bson:"seq"`
	}

	err := counterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("failed to get next ID: %w", err)
	}

	return counter.Seq, nil
}

// GetUserByGoogleID retrieves a user by Google ID
func (r *MongoRepository) GetUserByGoogleID(googleID string) (*storage.User, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("users")

	filter := bson.M{"google_id": googleID}

	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.bsonToUser(result), nil
}

// bsonToUser converts a BSON document to storage.User
func (r *MongoRepository) bsonToUser(doc bson.M) *storage.User {
	user := &storage.User{}

	if id, ok := doc["id"].(int32); ok {
		user.ID = int(id)
	} else if id, ok := doc["id"].(int64); ok {
		user.ID = int(id)
	} else if id, ok := doc["id"].(int); ok {
		user.ID = id
	}

	if googleID, ok := doc["google_id"].(string); ok {
		user.GoogleID = googleID
	}

	if email, ok := doc["email"].(string); ok {
		user.Email = email
	}

	if createdAt, ok := doc["created_at"].(time.Time); ok {
		user.CreatedAt = createdAt
	}

	if updatedAt, ok := doc["updated_at"].(time.Time); ok {
		user.UpdatedAt = updatedAt
	}

	return user
}

// CreateUser creates a new user
func (r *MongoRepository) CreateUser(googleID string, email string) (*storage.User, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("users")

	id, err := r.getNextID(ctx, "users")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := bson.M{
		"id":         id,
		"google_id":  googleID,
		"email":      email,
		"created_at": now,
		"updated_at": now,
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &storage.User{
		ID:        id,
		GoogleID:  googleID,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
