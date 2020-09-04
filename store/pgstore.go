package store

import (
	"database/sql"
	"fmt"

	// Load Postgres driver
	_ "github.com/lib/pq"
)

// PostgresConfig defines the parameters needed to initialize the
// Postgres connection.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBname   string
}

// PostgresStore if the type implmenting the Store interface for Postgres.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgres return a new Postgres store
func NewPostgres(cfg PostgresConfig) (*PostgresStore, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

// Memo gets a memoized fibonacci value
func (ps *PostgresStore) Memo(n int) (uint64, bool, error) {
	return 0, false, nil
}

// Memoize stores a memoized value
func (ps *PostgresStore) Memoize(v int, val uint64) error {
	return nil
}

// FindLessEqual finds the highest n and value memoized value less
// than or equal to the target
func (ps *PostgresStore) FindLessEqual(uint64) (int, uint64, error) {
	return 0, 0, nil
}

// Clear clears the table
func (ps *PostgresStore) Clear() error {
	return nil
}

// Shutdown shuts down the store
func (ps *PostgresStore) Shutdown() error {
	return ps.db.Close()

}
