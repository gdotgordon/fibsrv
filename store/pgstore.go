package store

// This is the implmentaton of the Postgres store.

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

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

	// The main queries are prepared at initialization time for efficiency.

	// query to select a memo by fibonacci number
	findMemo = `SELECT value FROM fibtab WHERE num = $1;`

	// count the number of mempoized artifacts where the value is less than
	// the target.
	memoCount = `SELECT count(*) as count FROM fibtab WHERE value < $1;`

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
	db        *sqlx.DB
	findStmt  *sqlx.Stmt
	storeStmt *sqlx.Stmt
	memoStmt  *sqlx.Stmt
	log       *zap.SugaredLogger
}

// Compile time interface implementation check.
var _ Store = (*PostgresStore)(nil)

// NewPostgres return a new Postgres store
func NewPostgres(ctx context.Context, cfg PostgresConfig, log *zap.SugaredLogger) (Store, error) {
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
			log.Debugw("Establishing DB connection", "status", "reached max connection attempts")
			return nil, err
		}
	}
	log.Infow("DB connection", "status", "connected to DB")

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

	store, err := db.PreparexContext(ctx, store)
	if err != nil {
		return nil, err
	}

	mcnt, err := db.PreparexContext(ctx, memoCount)
	if err != nil {
		return nil, err
	}

	ps := &PostgresStore{
		db:        db,
		findStmt:  find,
		storeStmt: store,
		memoStmt:  mcnt,
		log:       log,
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
	_, err := ps.storeStmt.ExecContext(ctx, n, val)
	return err
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
	var buf bytes.Buffer
	err1 := ps.findStmt.Close()
	if err1 != nil {
		buf.WriteString(err1.Error() + "\n")
	}
	err2 := ps.storeStmt.Close()
	if err2 != nil {
		buf.WriteString(err2.Error() + "\n")
	}
	err3 := ps.memoStmt.Close()
	if err3 != nil {
		buf.WriteString(err3.Error() + "\n")
	}
	err4 := ps.db.Close()
	if err4 != nil {
		buf.WriteString(err4.Error() + "\n")
	}
	if buf.Len() != 0 {
		return errors.New(buf.String())
	}
	return nil
}
