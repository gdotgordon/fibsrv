package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/gdotgordon/fibsrv/store"
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
	repo store.Store
)

// We use dockertest to set up a shared repo for the test.
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
		repo, err = store.NewPostgres(context.Background(),
			store.PostgresConfig{
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

// Tests finding the count of intermediate memos.
func TestFibLessDB(t *testing.T) {
	svc, err := NewFib(repo)
	if err != nil {
		t.Fatal(err)
	}
	err = svc.Clear(context.Background())
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	ctx := context.Background()
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range []struct {
		target uint64
		result int
	}{
		{target: 0, result: 0},
		{target: 1, result: 1},
		{target: 2, result: 3},
		{target: 11, result: 7},
		{target: 120, result: 12},
		{target: 11, result: 7},
		{target: 54, result: 10},
		{target: 21, result: 8},
	} {
		res, err := svc.FibLess(ctx, v.target)
		if err != nil {
			t.Fatal(err)
		}
		if res != v.result {
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.target, v.result, res)
		}
	}
}

// Tests running the fbonacci function with various values.
func TestFibDB(t *testing.T) {
	svc, err := NewFib(repo)
	if err != nil {
		t.Fatal(err)
	}
	err = svc.Clear(context.Background())
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	ctx := context.Background()
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range []struct {
		n      int
		result uint64
	}{
		{n: 0, result: 0},
		{n: 1, result: 1},
		{n: 2, result: 1},
		{n: 5, result: 5},
		{n: 10, result: 55},
		{n: 15, result: 610},
		{n: 20, result: 6765},
		{n: 8, result: 21},
	} {
		res, err := svc.Fib(ctx, v.n)
		if err != nil {
			t.Fatal(err)
		}
		if res != v.result {
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.result, v.n, res)
		}

		cnt, err := svc.MemoCount(ctx, v.result)
		if err != nil {
			t.Fatal(err)
		}

		if i == 2 {
			// The sequence starts: 0, 1, 1.  Therefore there is only one
			// memo less than 2 even though there are two previous memos.
			// So add one here for this special case.
			cnt++
		}
		if cnt != v.n {
			t.Fatalf("%d: expected %d memos, got %d", i, v.n, cnt)
		}
	}
}

// This benchmark calculates a range of Fibonacci numbers and clears the
// database after each calculation.  We expect this to be slow.
func BenchmarkFibonacciClearCache(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()
	svc, err := NewFib(repo)
	if err != nil {
		b.Fatal(err)
	}

	cnt := 15
	for n := 0; n < b.N; n++ {
		for i := 0; i < cnt; i++ {
			err = svc.Clear(context.Background())
			if err != nil {
				b.Fatalf("clear: %v", err)
			}
			f, err := svc.Fib(ctx, i)
			if err != nil {
				b.Fatalf("fib(%d): %v", i, err)
			}
			num, err := svc.MemoCount(ctx, f)
			if err != nil {
				b.Fatalf("les(%d): %v", num, err)
			}

			// Quirk of repeated 1's in Fibonacci sequence.
			if i == 2 {
				num++
			}
			if num != i {
				b.Fatalf("expcted %d memos, got %d, %d result", i, num, f)
			}
		}
	}
}

// This benchmark calculates a range of Fibonacci numbers but does not
// clear the database after each calculation.  We expect this to be faster.
func BenchmarkFibonacciNoClearCache(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()
	svc, err := NewFib(repo)
	if err != nil {
		b.Fatal(err)
	}

	err = svc.Clear(context.Background())
	if err != nil {
		b.Fatalf("clear: %v", err)
	}
	cnt := 15
	for n := 0; n < b.N; n++ {
		for i := 0; i < cnt; i++ {
			f, err := svc.Fib(ctx, i)
			if err != nil {
				b.Fatalf("fib(%d): %v", i, err)
			}
			num, err := svc.MemoCount(ctx, f)
			if err != nil {
				b.Fatalf("les(%d): %v", num, err)
			}
		}
	}
}

// This benchmark calculates a range of Fibonacci numbers and uses no
// caching and no database whatsover.  It turns out to be way faster
// than the other two because database.
func BenchmarkFibonacciNoCache(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()
	svc, err := NewFib(&NeverStore{vals: make(map[int]uint64)})
	if err != nil {
		b.Fatal(err)
	}

	err = svc.Clear(context.Background())
	if err != nil {
		b.Fatalf("clear: %v", err)
	}
	cnt := 15
	for n := 0; n < b.N; n++ {
		for i := 0; i < cnt; i++ {
			f, err := svc.Fib(ctx, i)
			if err != nil {
				b.Fatalf("fib(%d): %v", i, err)
			}
			// No memos being stored, ignore.
			svc.MemoCount(ctx, f)
		}
	}
}

// NeverStore is a store that defeats all caching, and is used
// for benchmarking.
type NeverStore struct {
	vals map[int]uint64
}

func (ns NeverStore) Memo(context.Context, int) (uint64, bool, error) {
	return 0, false, nil
}
func (ns NeverStore) Memoize(ctx context.Context, n int, value uint64) error {
	ns.vals[n] = value
	return nil
}
func (ns NeverStore) FindLess(context.Context, uint64) (*store.FibPair, error) {
	return nil, nil
}
func (ns NeverStore) MemoCount(context.Context, uint64) (int, error) {
	return len(ns.vals), nil
}
func (ns NeverStore) Clear(context.Context) error {
	ns.vals = make(map[int]uint64)
	return nil
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
