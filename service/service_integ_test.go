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

func TestFoo(t *testing.T) {
	svc, err := NewFib(repo)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("got service", svc)
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
		{target: 2, result: 3},
		{target: 11, result: 7},
		{target: 1, result: 1},
		{target: 120, result: 12},
	} {
		fmt.Println("fibless", v.target)
		res, err := svc.FibLess(ctx, v.target)
		if err != nil {
			t.Fatal(err)
		}
		if res != v.result {
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.target, v.result, res)
		}
	}
}
