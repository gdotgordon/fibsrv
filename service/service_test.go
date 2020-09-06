package service

import (
	"context"
	"testing"

	"github.com/gdotgordon/fibsrv/store"
)

// Testing fetching intermediate memo counts using a mock hash store.
// This allows us to focus on the business logic of the service.
func TestFibLess(t *testing.T) {
	ctx := context.Background()
	store := store.NewMap()
	svc, err := NewFib(store)
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
		res, err := svc.FibLess(ctx, v.target)
		if err != nil {
			t.Fatal(err)
		}
		if res != v.result {
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.target, v.result, res)
		}
	}
}

// Testing computing fibonacci values using a mock hash store.
// This allows us to focus on the business logic of the service.
func TestFib(t *testing.T) {
	ctx := context.Background()
	store := store.NewMap()
	svc, err := NewFib(store)
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
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.n, v.result, res)
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
