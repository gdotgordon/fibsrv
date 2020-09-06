// Package service defines the "business" logic, backed by
// a Store from the store package to manage the data.
// There only current implmentation is the FibService.
//
// The service package API's are invoked from the api package,
// which are the HTTP handler functions.
package service

import (
	"context"

	"github.com/gdotgordon/fibsrv/store"
)

// FibService defines the interface for the functions
type FibService interface {
	// Fib Finds fibonacci(n)
	Fib(context.Context, int) (uint64, error)

	// FibLess find the number of memoized values
	// less than the target value
	FibLess(context.Context, uint64) (int, error)

	// MemoCount counts the number of memoized values
	// less than or equal to a target
	MemoCount(context.Context, uint64) (int, error)

	// Clear emties all data from the store.
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
		//fmt.Println("hit:", n, val)
		return val, nil
	}
	if n == 0 {
		//fmt.Println("0 case")
		if err = fsi.store.Memoize(ctx, 0, 0); err != nil {
			return 0, err
		}
		return 0, nil
	}
	if n == 1 {
		//fmt.Println("1 case")
		if err = fsi.store.Memoize(ctx, 0, 0); err != nil {
			return 0, err
		}
		if err = fsi.store.Memoize(ctx, 1, 1); err != nil {
			return 0, err
		}
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

// FibLess finds the number of intermediate results such that their fib(n)
// is less than the target value.  It starts with the highest intermediate
// value in the database and then increments by 1 until it finds a value
// greater than or equal to the target.  At that point it goes back and fetches
// the number of memos less than the target value.
//
// In other words, we need to create any missing memos to get to the point
// that the repository has all the missing intermidate results.
//
// The logic depends on the fact that any memos populated to the db are
// done correctly, i.e. the complete sequence is recorded.  This will
// always be the case.
func (fsi *FibImpl) FibLess(ctx context.Context, target uint64) (int, error) {
	fp, err := fsi.store.FindLess(ctx, target)
	n := 0
	if err != nil {
		return 0, err
	}
	if fp != nil {
		// The highest memo found is equal to the target, so count the
		// number of intermediate results in range in the db and return that.
		if fp.Value == target {
			cnt, err := fsi.MemoCount(ctx, target)
			if err != nil {
				return 0, err
			}
			return cnt, nil
		}
		n = fp.Num
	}

	// We don't have enough memos in the db yet, so create the missing ones
	// and then count the memos stored.
	for {
		res, err := fsi.Fib(ctx, n+1)
		if err != nil {
			return 0, err
		}
		if res >= target {
			cnt, err := fsi.MemoCount(ctx, target)
			if err != nil {
				return 0, err
			}
			return cnt, nil
		}
		n++
	}
}

// MemoCount counts the number of memoizations less than or equal to
// the target.
func (fsi *FibImpl) MemoCount(ctx context.Context, target uint64) (int, error) {
	res, err := fsi.store.MemoCount(ctx, target)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// Clear clears all rows of the store.
func (fsi *FibImpl) Clear(ctx context.Context) error {
	return fsi.store.Clear(ctx)
}
