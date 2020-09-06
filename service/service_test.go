package service

import (
	"context"
	"testing"

	"github.com/gdotgordon/fibsrv/store"
)

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

func TestMemoCount(t *testing.T) {
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
		if err := svc.Clear(ctx); err != nil {
			t.Fatal(err)
		}
		res, err := svc.FibLess(ctx, v.target)
		if err != nil {
			t.Fatal(err)
		}
		if res != v.result {
			t.Fatalf("%d: memocnt(%d), expected %d, got %d", i, v.target, v.result, res)
		}
	}
}
