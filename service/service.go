package service

import (
	"context"
	"fmt"

	"github.com/gdotgordon/fibsrv/store"
)

// FibService defines the interface for the functions provided
type FibService interface {
	Fib(context.Context, int) (uint64, error)
	FibLess(context.Context, uint64) (int, error)
	Clear(context.Context) error
}

// FibImpl implements the fib service
type FibImpl struct {
	store store.Store
}

// NewFib returns a new Fibonacci service
func NewFib(store store.Store) (FibService, error) {
	return &FibImpl{store: store}, nil
}

// Fib gets the Fibonacci value for a number, returns an error if one occurs.
func (fsi *FibImpl) Fib(ctx context.Context, n int) (uint64, error) {
	val, ok, err := fsi.store.Memo(ctx, n)
	if err != nil {
		return 0, err
	}
	if ok {
		fmt.Println("hit:", n, val)
		return val, nil
	}
	if n == 0 {
		fmt.Println("0 case")
		fsi.store.Memoize(ctx, 0, 0)
		return 0, nil
	}
	if n == 1 {
		fmt.Println("1 case")
		fsi.store.Memoize(ctx, 1, 1)
		return 1, nil
	}
	f1, err := fsi.Fib(ctx, n-1)
	if err != nil {
		return 0, err
	}
	f2, err := fsi.Fib(ctx, n-2)
	if err != nil {
		return 0, err
	}
	res := f1 + f2
	if err := fsi.store.Memoize(ctx, n, res); err != nil {
		return 0, err
	}
	return res, nil
}

// FibLess finds Fibonacci(N) such that the value is the highest one less
// than the target value.  It starts with the highest intermediate value
// in the database and then increments by 1 until it finds a value greater
// than or equla to the target.
func (fsi *FibImpl) FibLess(ctx context.Context, target uint64) (int, error) {
	if target == 0 {
		return 0, nil
	}
	if target == 1 {
		return 1, nil
	}

	fp, err := fsi.store.FindLessEqual(ctx, target)
	var n int
	if err != nil {
		return 0, err
	}
	if fp != nil {
		if fp.Value == target {
			return fp.Num, nil
		}
		n = fp.Num
	} else {
		n = 0
	}

	for {
		res, err := fsi.Fib(ctx, n+1)
		if err != nil {
			return 0, err
		}
		if res >= target {
			return n + 1, nil
		}
		n++
	}
}

// Clear clears all rows of the store.
func (fsi *FibImpl) Clear(ctx context.Context) error {
	return fsi.store.Clear(ctx)
}
