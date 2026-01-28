package storage

import "database/sql"

// DatabaseInterface defines the interface that both Postgres and MongoDB implementations must satisfy
// This allows repositories to work with either database type
type DatabaseInterface interface {
	// GetSQLDB returns the underlying *sql.DB for Postgres repositories
	// Returns nil for MongoDB implementations
	GetSQLDB() *sql.DB
	
	// GetMongoClient returns the underlying MongoDB client
	// Returns nil for Postgres implementations
	GetMongoClient() interface{}
	
	// RunMigrations sets up the database schema/indexes
	RunMigrations() error
	
	// Close closes the database connection
	Close() error
}
