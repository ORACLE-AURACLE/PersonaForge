package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// PostgresDatabase wraps the sql.DB connection for PostgreSQL
type PostgresDatabase struct {
	DB *sql.DB
}

// NewPostgresDatabase creates a new PostgreSQL database connection
func NewPostgresDatabase(databaseURL string) (*PostgresDatabase, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDatabase{DB: db}, nil
}

// GetSQLDB returns the underlying *sql.DB
func (d *PostgresDatabase) GetSQLDB() *sql.DB {
	return d.DB
}

// GetMongoClient returns nil for Postgres
func (d *PostgresDatabase) GetMongoClient() interface{} {
	return nil
}

// RunMigrations executes all pending database migrations
func (d *PostgresDatabase) RunMigrations() error {
	// Ensure migrations table exists
	_, err := d.DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedVersions, err := d.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files
	var migrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrations = append(migrations, entry.Name())
		}
	}
	sort.Strings(migrations)

	// Apply pending migrations
	for _, filename := range migrations {
		version := extractVersion(filename)
		if version == 0 {
			continue
		}

		// Skip if already applied
		if _, applied := appliedVersions[version]; applied {
			continue
		}

		// Read migration file
		content, err := migrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration
		if err := d.executeMigration(version, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		fmt.Printf("Applied migration: %s\n", filename)
	}

	return nil
}

func (d *PostgresDatabase) getAppliedMigrations() (map[int]bool, error) {
	rows, err := d.DB.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func (d *PostgresDatabase) executeMigration(version int, sql string) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Execute migration SQL
	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	// Record migration (only if not already in the SQL)
	if !strings.Contains(sql, "INSERT INTO schema_migrations") {
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func extractVersion(filename string) int {
	// Extract version from filename like "001_initial_schema.sql"
	parts := strings.Split(filename, "_")
	if len(parts) == 0 {
		return 0
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	return version
}

// Close closes the database connection
func (d *PostgresDatabase) Close() error {
	return d.DB.Close()
}
