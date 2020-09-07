package service

// Benchmark tests using Postgres backing store.

import (
	"context"
	"testing"
)

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
			_, err := svc.Fib(ctx, i)
			if err != nil {
				b.Fatalf("fib(%d): %v", i, err)
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
			_, err := svc.Fib(ctx, i)
			if err != nil {
				b.Fatalf("fib(%d): %v", i, err)
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
			_, err := svc.Fib(ctx, i)
			if err != nil {
				b.Fatalf("fib(%d): %v", i, err)
			}
		}
	}
}
