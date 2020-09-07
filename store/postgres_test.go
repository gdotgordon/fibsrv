// +build store

// Run as: go test -tags=store
package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"go.uber.org/zap"
)

var (
	user     = "postgres"
	password = "secret"
	db       = "fib_db"
	port     = "5433"
	dialect  = "postgres"
	dsn      = "postgres://%s:%s@localhost:%s/%s?sslmode=disable"
	idleConn = 25
	maxConn  = 25
)

var (
	repo Store
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + db,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	resource, err := pool.RunWithOptions(&opts)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err.Error())
	}

	dsn = fmt.Sprintf(dsn, user, password, port, db)
	portNum, _ := strconv.Atoi(port)
	if err = pool.Retry(func() error {
		repo, err = NewPostgres(context.Background(),
			PostgresConfig{
				Host:     "localhost",
				Port:     portNum,
				User:     user,
				Password: password,
				DBName:   db,
			},
			newDebugLogger(),
		)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err.Error())
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

// Exercise all the opreations supported by the store.
// TODO: add more test cases, but this is the basic idea
// to utilize all the APIs.
func TestStore(t *testing.T) {

	ctx := context.Background()
	if err := repo.Clear(ctx); err != nil {
		t.Fatalf("error clearing store: %v", err)
	}

	for i, v := range []struct {
		pairs       []FibPair
		missingKeys []int
	}{
		{
			pairs:       []FibPair{{0, 0}, {1, 1}, {2, 1}, {3, 2}, {4, 3}},
			missingKeys: []int{5, 7},
		},
	} {
		// Memoize the values.
		for _, p := range v.pairs {
			if err := repo.Memoize(ctx, p.Num, p.Value); err != nil {
				t.Fatalf("%d: error memoizing %v: %v", i, p, err)
			}
		}

		// Re-Memoizing the values should be a harmless no-op
		for _, p := range v.pairs {
			if err := repo.Memoize(ctx, p.Num, p.Value); err != nil {
				t.Fatalf("%d: error memoizing %v: %v", i, p, err)
			}
		}

		// Get the memos we srtored back.
		var lastVal uint64
		for i, p := range v.pairs {
			val, ok, err := repo.Memo(ctx, p.Num)
			if err != nil {
				t.Fatalf("%d: error retrieving memo %d: %v", i, p.Num, err)
			}
			if !ok {
				t.Fatalf("%d: couldn't retrieve memo %d: %v", i, p.Num, err)
			}
			if val != p.Value {
				t.Fatalf("%d: retrieving memo %d, expected value %d, got %d", i, p, p.Value, val)
			}
			if i == len(v.pairs)-1 {
				lastVal = p.Value
			}
		}

		// Make sure we handle missng memos ok.
		for _, m := range v.missingKeys {
			_, ok, err := repo.Memo(ctx, m)
			if err != nil {
				t.Fatalf("%d: error retrieving missing memo %d: %v", i, m, err)
			}
			if ok {
				t.Fatalf("%d: did not expect to get memo %d", i, m)
			}
		}

		// Verify the memo count.
		cnt, err := repo.MemoCount(ctx, lastVal)
		if err != nil {
			t.Fatalf("%d: error getting memo count: %v", i, err)
		}
		if cnt != len(v.pairs)-1 {
			t.Fatalf("%d: retrieving memo coiunt, expected value %d, got %d", i, len(v.pairs)-1, cnt)
		}

		if err := repo.Clear(ctx); err != nil {
			t.Fatalf("%d: error clearing store: %v", i, err)
		}

		// And again after clearing the repo.
		cnt, err = repo.MemoCount(ctx, lastVal)
		if err != nil {
			t.Fatalf("%d: error getting memo count: %v", i, err)
		}
		if cnt != 0 {
			t.Fatalf("%d: retrieving memo coiunt, expected value 0, got %d", i, cnt)
		}
	}
}

func newDebugLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	lg, _ := config.Build()
	return lg.Sugar()
}

func newNoopLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}
	lg, _ := config.Build()
	return lg.Sugar()
}
