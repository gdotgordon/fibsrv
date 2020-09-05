package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// Load Postgres driver
	_ "github.com/lib/pq"
)

const (
	createTable = `CREATE TABLE IF NOT EXISTS fibtab (
	num INTEGER PRIMARY KEY,
	value BIGINT
	);`

	findMemo = `SELECT value FROM fibtab WHERE num = $1;`

	findLess = `SELECT num, value FROM fibtab WHERE value <= $1 ORDER BY VALUE DESC LIMIT 1;`

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

// PostgresStore if the type implmenting the Store interface for Postgres.
type PostgresStore struct {
	db           *sqlx.DB
	findStmt     *sqlx.Stmt
	findLessStmt *sqlx.Stmt
	storeStmt    *sqlx.Stmt
}

// NewPostgres return a new Postgres store
func NewPostgres(ctx context.Context, cfg PostgresConfig) (*PostgresStore, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	fmt.Println("****connecting to db!", psqlInfo)
	var db *sqlx.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sqlx.ConnectContext(ctx, "postgres", psqlInfo)
		if err != nil {
			fmt.Println("****got err!", err)
			time.Sleep(1 * time.Second)
		}
		if i == 10 {
			fmt.Println("too many errors")
			return nil, err
		}
	}
	fmt.Println("****connected to db!")

	_, err = db.ExecContext(ctx, createTable)
	if err != nil {
		return nil, err
	}

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

	ps := &PostgresStore{
		db:           db,
		findStmt:     find,
		findLessStmt: findLess,
		storeStmt:    store,
	}
	return ps, nil
}

// Memo gets a memoized fibonacci value
func (ps *PostgresStore) Memo(ctx context.Context, n int) (uint64, bool, error) {
	fmt.Println("get memo", n)
	var res []uint64
	err := ps.findStmt.SelectContext(ctx, &res, n)
	if err != nil {
		return 0, false, err
	}
	if len(res) != 1 {
		fmt.Println("memo not found", n)
		return 0, false, nil
	}
	fmt.Println("memo", n, res[0])
	return res[0], true, nil
}

// Memoize stores a memoized value
func (ps *PostgresStore) Memoize(ctx context.Context, n int, val uint64) error {
	fmt.Println("memoize", n, val)
	_, err := ps.storeStmt.ExecContext(ctx, n, val)
	return err
}

// FindLessEqual finds the highest n and value memoized value less
// than or equal to the target
func (ps *PostgresStore) FindLessEqual(ctx context.Context, target uint64) (*FibPair, error) {
	var fp []FibPair
	err := ps.findLessStmt.SelectContext(ctx, &fp, target)
	if err != nil {
		return nil, err
	}
	if len(fp) != 1 {
		fmt.Println("******le nothing found!!!!")
		return nil, nil
	}
	fmt.Println("******le got", &fp[0])
	return &fp[0], nil
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
