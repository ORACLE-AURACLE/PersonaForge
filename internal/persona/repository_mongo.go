package persona

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/PersonaForge/backend/internal/storage"
)

// SessionActive checks whether a session exists and is not expired.
func (r *MongoRepository) SessionActive(sessionID string) (bool, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("sessions")

	filter := bson.M{
		"session_id": sessionID,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check session: %w", err)
	}

	return count > 0, nil
}

// getNextID gets the next sequential ID for a collection
func (r *MongoRepository) getNextID(ctx context.Context, collectionName string) (int, error) {
	counterCollection := r.db.Database.Collection("counters")

	filter := bson.M{"_id": collectionName}
	update := bson.M{
		"$inc": bson.M{"seq": 1},
		// "$setOnInsert": bson.M{"seq": 0},
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

// CreatePersona creates a new custom persona
func (r *MongoRepository) CreatePersona(userID *int, sessionID *string, name string, blueprint string) (*storage.Persona, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	// Get next ID
	id, err := r.getNextID(ctx, "personas")
	if err != nil {
		return nil, err
	}

	persona := bson.M{
		"id":         id,
		"user_id":    userID,
		"session_id": sessionID,
		"name":       name,
		"blueprint":  blueprint,
		"is_default": false,
		"created_at": time.Now(),
	}

	_, err = collection.InsertOne(ctx, persona)
	if err != nil {
		return nil, fmt.Errorf("failed to create persona: %w", err)
	}

	// Return the created persona
	result := &storage.Persona{
		ID:        id,
		UserID:    userID,
		SessionID: sessionID,
		Name:      name,
		Blueprint: blueprint,
		IsDefault: false,
		CreatedAt: persona["created_at"].(time.Time),
	}

	return result, nil
}

// GetPersonaByID retrieves a persona by ID
func (r *MongoRepository) GetPersonaByID(id int) (*storage.Persona, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	filter := bson.M{"id": id}

	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("persona not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get persona: %w", err)
	}

	return r.bsonToPersona(result), nil
}

// bsonToPersona converts a BSON document to storage.Persona
func (r *MongoRepository) bsonToPersona(doc bson.M) *storage.Persona {
	persona := &storage.Persona{}

	if id, ok := doc["id"].(int32); ok {
		persona.ID = int(id)
	} else if id, ok := doc["id"].(int64); ok {
		persona.ID = int(id)
	} else if id, ok := doc["id"].(int); ok {
		persona.ID = id
	}

	if userID, ok := doc["user_id"].(int32); ok {
		uid := int(userID)
		persona.UserID = &uid
	} else if userID, ok := doc["user_id"].(int64); ok {
		uid := int(userID)
		persona.UserID = &uid
	} else if userID, ok := doc["user_id"].(int); ok {
		persona.UserID = &userID
	} else if doc["user_id"] == nil {
		persona.UserID = nil
	}

	if sessionID, ok := doc["session_id"].(string); ok {
		persona.SessionID = &sessionID
	} else if doc["session_id"] == nil {
		persona.SessionID = nil
	}

	if name, ok := doc["name"].(string); ok {
		persona.Name = name
	}

	if blueprint, ok := doc["blueprint"].(string); ok {
		persona.Blueprint = blueprint
	}

	if isDefault, ok := doc["is_default"].(bool); ok {
		persona.IsDefault = isDefault
	}

	if createdAt, ok := doc["created_at"].(time.Time); ok {
		persona.CreatedAt = createdAt
	}

	return persona
}

// ListDefaultPersonas retrieves all default personas.
func (r *MongoRepository) ListDefaultPersonas() ([]storage.Persona, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	filter := bson.M{"is_default": true}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode personas: %w", err)
	}

	personas := make([]storage.Persona, 0, len(results))
	for _, doc := range results {
		personas = append(personas, *r.bsonToPersona(doc))
	}

	return personas, nil
}

// ListPersonasForUser retrieves all personas for a user (including defaults)
func (r *MongoRepository) ListPersonasForUser(userID int) ([]storage.Persona, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	filter := bson.M{
		"$or": []bson.M{
			{"is_default": true},
			{"user_id": userID},
		},
	}
	opts := options.Find().SetSort(bson.D{
		{Key: "is_default", Value: -1},
		{Key: "created_at", Value: -1},
	})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode personas: %w", err)
	}

	personas := make([]storage.Persona, 0, len(results))
	for _, doc := range results {
		personas = append(personas, *r.bsonToPersona(doc))
	}

	return personas, nil
}

// ListPersonasForSession retrieves all personas for a guest session (including defaults)
func (r *MongoRepository) ListPersonasForSession(sessionID string) ([]storage.Persona, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	filter := bson.M{
		"$or": []bson.M{
			{"is_default": true},
			{"session_id": sessionID},
		},
	}
	opts := options.Find().SetSort(bson.D{
		{Key: "is_default", Value: -1},
		{Key: "created_at", Value: -1},
	})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode personas: %w", err)
	}

	personas := make([]storage.Persona, 0, len(results))
	for _, doc := range results {
		personas = append(personas, *r.bsonToPersona(doc))
	}

	return personas, nil
}

// ListCustomPersonasForSession returns only custom personas created by the given session.
func (r *MongoRepository) ListCustomPersonasForSession(sessionID string) ([]storage.Persona, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")
	filter := bson.M{"session_id": sessionID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas for session: %w", err)
	}
	defer cursor.Close(ctx)
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode personas: %w", err)
	}
	personas := make([]storage.Persona, 0, len(results))
	for _, doc := range results {
		personas = append(personas, *r.bsonToPersona(doc))
	}
	return personas, nil
}

// CountCustomPersonasForUser counts non-default personas for a user
func (r *MongoRepository) CountCustomPersonasForUser(userID int) (int, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	filter := bson.M{
		"user_id":    userID,
		"is_default": false,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count personas: %w", err)
	}

	return int(count), nil
}

// CountCustomPersonasForSession counts non-default personas for a guest session
func (r *MongoRepository) CountCustomPersonasForSession(sessionID string) (int, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	filter := bson.M{
		"session_id": sessionID,
		"is_default": false,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count personas: %w", err)
	}

	return int(count), nil
}

// DeletePersona deletes a custom persona
func (r *MongoRepository) DeletePersona(id int, userID int) error {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	filter := bson.M{
		"id":         id,
		"user_id":    userID,
		"is_default": false,
	}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete persona: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("persona not found or cannot be deleted")
	}

	return nil
}

// InitializeDefaultPersonas creates the 4 default personas if they don't exist
func (r *MongoRepository) InitializeDefaultPersonas() error {
	ctx := context.TODO()
	collection := r.db.Database.Collection("personas")

	// Check if defaults already exist
	count, err := collection.CountDocuments(ctx, bson.M{"is_default": true})
	if err != nil {
		return fmt.Errorf("failed to check default personas: %w", err)
	}

	if count > 0 {
		return nil // Already initialized
	}

	// Create default personas
	defaults := DefaultPersonas()
	for _, blueprint := range defaults {
		blueprintJSON, err := MarshalBlueprint(blueprint)
		if err != nil {
			return fmt.Errorf("failed to marshal blueprint: %w", err)
		}

		// Get next ID for this persona
		id, err := r.getNextID(ctx, "personas")
		if err != nil {
			return fmt.Errorf("failed to get next ID for default persona: %w", err)
		}

		persona := bson.M{
			"id":         id,
			"user_id":    nil,
			"session_id": nil,
			"name":       blueprint.Name,
			"blueprint":  blueprintJSON,
			"is_default": true,
			"created_at": time.Now(),
		}

		if _, err := collection.InsertOne(ctx, persona); err != nil {
			return fmt.Errorf("failed to create default persona: %w", err)
		}
	}

	return nil
}
