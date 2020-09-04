package service

import (
	"testing"

	"github.com/gdotgordon/fibsrv/store"
)

func TestFibLess(t *testing.T) {
	store := store.NewMap()
	svc, err := NewFib(store)
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range []struct {
		target uint64
		result int
	}{
		{target: 11, result: 4},
	} {
		res, err := svc.FibLess(v.target)
		if err != nil {
			t.Fatal(err)
		}
		if res != v.result {
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.target, v.result, res)
		}
	}
}
