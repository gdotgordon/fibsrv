package service

import (
	"fmt"

	"github.com/gdotgordon/fibsrv/store"
)

// FibService defines the interface for the functions provided
type FibService interface {
	Fib(int) (uint64, error)
	FibLess(uint64) (int, error)
	Clear() error
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
func (fsi *FibImpl) Fib(n int) (uint64, error) {
	val, ok, err := fsi.store.Memo(n)
	if err != nil {
		return 0, err
	}
	if ok {
		fmt.Println("hit:", n, val)
		return val, nil
	}
	if n == 0 {
		fmt.Println("0 case")
		fsi.store.Memoize(0, 0)
		return 0, nil
	}
	if n == 1 {
		fmt.Println("1 case")
		fsi.store.Memoize(1, 1)
		return 1, nil
	}
	f1, err := fsi.Fib(n - 1)
	if err != nil {
		return 0, err
	}
	f2, err := fsi.Fib(n - 2)
	if err != nil {
		return 0, err
	}
	res := f1 + f2
	if err := fsi.store.Memoize(n, res); err != nil {
		return 0, err
	}
	if err := fsi.store.Memoize(n, res); err != nil {
		return 0, err
	}
	return res, nil
}

// FibLess finds Fibonacci(N) such that the value is the highest one less
// than the target value.
func (fsi *FibImpl) FibLess(target uint64) (int, error) {
	if target == 0 {
		return 0, nil
	}

	n, val, err := fsi.store.FindLessEqual(target)
	if err != nil {
		return 0, err
	}
	if val == target {
		return n, nil
	}

	fmt.Println("intermediate:", n, val)
	for {
		res, err := fsi.Fib(n + 1)
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
func (fsi *FibImpl) Clear() error {
	return fsi.store.Clear()
}
