package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// Load Postgres driver
	_ "github.com/lib/pq"
)

const (
	// the table has the "n" of fib(n) as the primary key, and the value
	// of fib(n) stored as a big integer.
	createTable = `CREATE TABLE IF NOT EXISTS fibtab (
	num INTEGER PRIMARY KEY,
	value BIGINT
	);`

	// query to select a memo by fibonacci number
	findMemo = `SELECT value FROM fibtab WHERE num = $1;`

	// finds the highest memoized number and value where the value is
	// less than the target.
	findLess = `SELECT num, value FROM fibtab WHERE value <= $1 ORDER BY VALUE DESC LIMIT 1;`

	// count the number of mempoized artifacts where the value is less than
	// or equal to the target.
	memoCount = `SELECT count(*) as count FROM fibtab WHERE value <= $1;`

	// store a memo (ignore a duplicate update)
	store = `INSERT INTO fibtab (num, value) VALUES ($1, $2)
	    ON CONFLICT (num) DO NOTHING;`
)

// PostgresConfig defines the parameters needed to initialize the
// Postgres connection.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// PostgresStore if the type implementing the Store interface for Postgres.
type PostgresStore struct {
	db           *sqlx.DB
	findStmt     *sqlx.Stmt
	findLessStmt *sqlx.Stmt
	storeStmt    *sqlx.Stmt
	memoStmt     *sqlx.Stmt
}

// NewPostgres return a new Postgres store
func NewPostgres(ctx context.Context, cfg PostgresConfig) (Store, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	var db *sqlx.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sqlx.ConnectContext(ctx, "postgres", psqlInfo)
		if err != nil {
			time.Sleep(1 * time.Second)
		}
		if i == 10 {
			fmt.Println("too many errors")
			return nil, err
		}
	}
	fmt.Println("****connected to db!")

	// Create the table if it doesn't exist.
	_, err = db.ExecContext(ctx, createTable)
	if err != nil {
		return nil, err
	}

	// Prepare the other statements for better performance.
	find, err := db.PreparexContext(ctx, findMemo)
	if err != nil {
		return nil, err
	}

	findLess, err := db.PreparexContext(ctx, findLess)
	if err != nil {
		return nil, err
	}

	store, err := db.PreparexContext(ctx, store)
	if err != nil {
		return nil, err
	}

	mcnt, err := db.PreparexContext(ctx, memoCount)
	if err != nil {
		return nil, err
	}

	ps := &PostgresStore{
		db:           db,
		findStmt:     find,
		findLessStmt: findLess,
		storeStmt:    store,
		memoStmt:     mcnt,
	}
	return ps, nil
}

// Memo gets a memoized fibonacci value.  Returns false for the second
// parameter if not yet cached.
func (ps *PostgresStore) Memo(ctx context.Context, n int) (uint64, bool, error) {
	var res uint64
	row := ps.findStmt.QueryRowContext(ctx, n)
	err := row.Scan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return res, true, nil
}

// Memoize stores a memoized value
func (ps *PostgresStore) Memoize(ctx context.Context, n int, val uint64) error {
	//fmt.Println("memoize", n, val)
	_, err := ps.storeStmt.ExecContext(ctx, n, val)
	return err
}

// FindLessEqual finds the highest n and value memoized value less
// than or equal to the target
func (ps *PostgresStore) FindLessEqual(ctx context.Context, target uint64) (*FibPair, error) {
	var fp FibPair
	row := ps.findLessStmt.QueryRowContext(ctx, target)
	err := row.Scan(&fp.Num, &fp.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &fp, nil
}

// MemoCount returns the number of memoizations whose value is less than or
// equal to the target.
func (ps *PostgresStore) MemoCount(ctx context.Context, target uint64) (int, error) {
	var res int
	row := ps.memoStmt.QueryRowContext(ctx, target)
	err := row.Scan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return res, nil
}

// Clear clears the table
func (ps *PostgresStore) Clear(ctx context.Context) error {
	_, err := ps.db.ExecContext(ctx, "TRUNCATE fibtab;")
	return err
}

// Shutdown shuts down the store
func (ps *PostgresStore) Shutdown() error {
	return ps.db.Close()

}
