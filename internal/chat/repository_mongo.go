package chat

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/PersonaForge/backend/internal/storage"
)

// MongoRepository handles chat data access for MongoDB
type MongoRepository struct {
	db *storage.MongoDatabase
}

// NewMongoRepository creates a new chat repository for MongoDB
func NewMongoRepository(db *storage.MongoDatabase) *MongoRepository {
	return &MongoRepository{db: db}
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

// CreateSession creates a new chat session
func (r *MongoRepository) CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("sessions")

	id, err := r.getNextID(ctx, "sessions")
	if err != nil {
		return 0, err
	}

	session := bson.M{
		"id":           id,
		"user_id":      userID,
		"session_id":   sessionID,
		"is_anonymous": isAnonymous,
		"created_at":   time.Now(),
		"expires_at":   expiresAt,
	}

	fmt.Printf("[DEBUG] CreateSession - Creating session_id: %s in database: %s\n", sessionID, r.db.Database.Name())

	_, err = collection.InsertOne(ctx, session)
	if err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Printf("[DEBUG] CreateSession - Session created successfully with id: %d\n", id)
	return id, nil
}

// GetSessionByID retrieves a session by session ID string
func (r *MongoRepository) GetSessionByID(sessionID string) (*storage.Session, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("sessions")

	filter := bson.M{"session_id": sessionID}

	// Debug logging
	fmt.Printf("[DEBUG] GetSessionByID - Looking for session_id: %s\n", sessionID)
	fmt.Printf("[DEBUG] GetSessionByID - Database: %s, Collection: sessions\n", r.db.Database.Name())

	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("[DEBUG] GetSessionByID - No documents found for session_id: %s\n", sessionID)

		// List all sessions to debug
		cursor, listErr := collection.Find(ctx, bson.M{})
		if listErr == nil {
			var allSessions []bson.M
			if cursor.All(ctx, &allSessions) == nil {
				fmt.Printf("[DEBUG] GetSessionByID - Total sessions in DB: %d\n", len(allSessions))
				for i, s := range allSessions {
					if i < 3 { // Show first 3 sessions
						fmt.Printf("[DEBUG] GetSessionByID - Sample session %d: session_id=%v\n", i+1, s["session_id"])
					}
				}
			}
			cursor.Close(ctx)
		}

		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	fmt.Printf("[DEBUG] GetSessionByID - Session found: %+v\n", result)
	return r.bsonToSession(result), nil
}

// bsonToSession converts a BSON document to storage.Session
func (r *MongoRepository) bsonToSession(doc bson.M) *storage.Session {
	session := &storage.Session{}

	if id, ok := doc["id"].(int32); ok {
		session.ID = int(id)
	} else if id, ok := doc["id"].(int64); ok {
		session.ID = int(id)
	} else if id, ok := doc["id"].(int); ok {
		session.ID = id
	}

	if userID, ok := doc["user_id"].(int32); ok {
		uid := int(userID)
		session.UserID = &uid
	} else if userID, ok := doc["user_id"].(int64); ok {
		uid := int(userID)
		session.UserID = &uid
	} else if userID, ok := doc["user_id"].(int); ok {
		session.UserID = &userID
	} else if doc["user_id"] == nil {
		session.UserID = nil
	}

	if sessionID, ok := doc["session_id"].(string); ok {
		session.SessionID = sessionID
	}

	if isAnonymous, ok := doc["is_anonymous"].(bool); ok {
		session.IsAnonymous = isAnonymous
	}

	if createdAt, ok := doc["created_at"].(time.Time); ok {
		session.CreatedAt = createdAt
	} else if createdAtPrimitive, ok := doc["created_at"].(bson.DateTime); ok {
		session.CreatedAt = createdAtPrimitive.Time()
	}

	if expiresAt, ok := doc["expires_at"].(time.Time); ok {
		session.ExpiresAt = expiresAt
	} else if expiresAtPrimitive, ok := doc["expires_at"].(bson.DateTime); ok {
		session.ExpiresAt = expiresAtPrimitive.Time()
	}

	return session
}

// MigrateSession attaches an anonymous session to a user
func (r *MongoRepository) MigrateSession(sessionID string, userID int) error {
	ctx := context.TODO()
	collection := r.db.Database.Collection("sessions")

	filter := bson.M{
		"session_id":   sessionID,
		"is_anonymous": true,
	}
	update := bson.M{
		"$set": bson.M{
			"user_id":      userID,
			"is_anonymous": false,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to migrate session: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("session not found or not anonymous")
	}

	return nil
}

// SaveMessage saves a chat message
func (r *MongoRepository) SaveMessage(sessionDBID int, personaID int, role string, content string) (*MessageDTO, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("messages")

	id, err := r.getNextID(ctx, "messages")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	message := bson.M{
		"id":         id,
		"session_id": sessionDBID,
		"persona_id": personaID,
		"role":       role,
		"content":    content,
		"created_at": now,
	}

	_, err = collection.InsertOne(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return &MessageDTO{
		ID:        id,
		SessionID: sessionDBID,
		PersonaID: personaID,
		Role:      role,
		Content:   content,
		CreatedAt: now,
	}, nil
}

// GetConversationHistory retrieves all messages for a session
func (r *MongoRepository) GetConversationHistory(sessionDBID int, personaID int) ([]MessageDTO, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("messages")

	filter := bson.M{
		"session_id": sessionDBID,
		"persona_id": personaID,
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}

	messages := make([]MessageDTO, 0, len(results))
	for _, doc := range results {
		msg := MessageDTO{}

		if id, ok := doc["id"].(int32); ok {
			msg.ID = int(id)
		} else if id, ok := doc["id"].(int64); ok {
			msg.ID = int(id)
		} else if id, ok := doc["id"].(int); ok {
			msg.ID = id
		}

		if sessionID, ok := doc["session_id"].(int32); ok {
			msg.SessionID = int(sessionID)
		} else if sessionID, ok := doc["session_id"].(int64); ok {
			msg.SessionID = int(sessionID)
		} else if sessionID, ok := doc["session_id"].(int); ok {
			msg.SessionID = sessionID
		}

		if personaID, ok := doc["persona_id"].(int32); ok {
			msg.PersonaID = int(personaID)
		} else if personaID, ok := doc["persona_id"].(int64); ok {
			msg.PersonaID = int(personaID)
		} else if personaID, ok := doc["persona_id"].(int); ok {
			msg.PersonaID = personaID
		}

		if role, ok := doc["role"].(string); ok {
			msg.Role = role
		}

		if content, ok := doc["content"].(string); ok {
			msg.Content = content
		}

		if createdAt, ok := doc["created_at"].(time.Time); ok {
			msg.CreatedAt = createdAt
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// GetAllMessagesForSession retrieves all messages for a session across all personas.
func (r *MongoRepository) GetAllMessagesForSession(sessionDBID int) ([]MessageDTO, error) {
	ctx := context.TODO()
	collection := r.db.Database.Collection("messages")

	filter := bson.M{"session_id": sessionDBID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}

	messages := make([]MessageDTO, 0, len(results))
	for _, doc := range results {
		msg := MessageDTO{}

		if id, ok := doc["id"].(int32); ok {
			msg.ID = int(id)
		} else if id, ok := doc["id"].(int64); ok {
			msg.ID = int(id)
		} else if id, ok := doc["id"].(int); ok {
			msg.ID = id
		}

		if sessionID, ok := doc["session_id"].(int32); ok {
			msg.SessionID = int(sessionID)
		} else if sessionID, ok := doc["session_id"].(int64); ok {
			msg.SessionID = int(sessionID)
		} else if sessionID, ok := doc["session_id"].(int); ok {
			msg.SessionID = sessionID
		}

		if personaID, ok := doc["persona_id"].(int32); ok {
			msg.PersonaID = int(personaID)
		} else if personaID, ok := doc["persona_id"].(int64); ok {
			msg.PersonaID = int(personaID)
		} else if personaID, ok := doc["persona_id"].(int); ok {
			msg.PersonaID = personaID
		}

		if role, ok := doc["role"].(string); ok {
			msg.Role = role
		}

		if content, ok := doc["content"].(string); ok {
			msg.Content = content
		}

		if createdAt, ok := doc["created_at"].(time.Time); ok {
			msg.CreatedAt = createdAt
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// SaveTokenUsage records token usage for a session
func (r *MongoRepository) SaveTokenUsage(sessionDBID int, promptTokens int, completionTokens int, totalTokens int) error {
	ctx := context.TODO()
	collection := r.db.Database.Collection("token_usage")

	tokenUsage := bson.M{
		"session_id":        sessionDBID,
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      totalTokens,
		"created_at":        time.Now(),
	}

	_, err := collection.InsertOne(ctx, tokenUsage)
	if err != nil {
		return fmt.Errorf("failed to save token usage: %w", err)
	}

	return nil
}
