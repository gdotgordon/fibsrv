// Package service defines the "business" logic, backed by
// a Store from the store package to manage the data.
// There only current implmentation is the FibService.
//
// The service package API's are invoked from the api package,
// which are the HTTP handler functions.
package service

import (
	"context"
	"fmt"

	"github.com/gdotgordon/fibsrv/store"
)

// FibService implements the fib service
type FibService struct {
	store store.Store
}

// NewFib returns a new Fibonacci service
func NewFib(store store.Store) (*FibService, error) {
	return &FibService{store: store}, nil
}

// Fib gets the Fibonacci value for a number, returns an error if one occurs.
// This is the standard recursive algorithm that also memoizes every
// fib(n) computation.  Thus, after a call to fib(n), we can expect
// a database entry to exist for every value 0 to n.  Some may have been
// prsnt from before, some not.
func (fs *FibService) Fib(ctx context.Context, n int) (uint64, error) {
	if n < 0 {
		return 0, fmt.Errorf("invalid fibonacci request: %d", n)
	}

	val, ok, err := fs.store.Memo(ctx, n)
	if err != nil {
		return 0, err
	}
	if ok {
		return val, nil
	}
	if n == 0 {
		if err = fs.store.Memoize(ctx, 0, 0); err != nil {
			return 0, err
		}
		return 0, nil
	}
	if n == 1 {
		if err = fs.store.Memoize(ctx, 1, 1); err != nil {
			return 0, err
		}
		return 1, nil
	}

	// This looks somewhat different than the standard fibonacci
	// recursion due to having to check for errors.
	f1, err := fs.Fib(ctx, n-1)
	if err != nil {
		return 0, err
	}
	f2, err := fs.Fib(ctx, n-2)
	if err != nil {
		return 0, err
	}
	res := f1 + f2
	if err := fs.store.Memoize(ctx, n, res); err != nil {
		return 0, err
	}
	return res, nil
}

// FibLess finds the number of intermediate results such that their fib(n)
// is less than the target value.  It basically keeps computing f(n) for
// larger and larger n, and finally when it is large enough, it returns
// the number of stored memos less than the target value.
//
// Note, if the target is an exact fibonacci number, e.g. f(10)=34, then this will
// return the number of *intermediate* memos, not incuding the memo for 34
// itself, so 9 for 34.
//
// Also note there are more efficient ways to do this if all the consecutive memos
// are stored, and our code does build a complete set of memos for each value computed.
// However, due to errors or other changes in the database, it makes the code too
// fragile to depend on this.
func (fs *FibService) FibLess(ctx context.Context, target uint64) (int, error) {

	// Keep computing fib(n) until we have a big enough value.  If
	// the cache is well populated, this should perform well.
	for n := 0; ; n++ {
		res, err := fs.Fib(ctx, n)
		if err != nil {
			return 0, err
		}
		if res >= target {
			cnt, err := fs.store.MemoCount(ctx, target)
			if err != nil {
				return 0, err
			}
			return cnt, nil
		}
	}
}

// Clear clears all rows of the store.
func (fs *FibService) Clear(ctx context.Context) error {
	return fs.store.Clear(ctx)
}
