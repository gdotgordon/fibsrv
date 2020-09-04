package store

import "sync"

// MapStore is a hash map implementation of the store
type MapStore struct {
	tab map[int]uint64
	mu  sync.Mutex
}

// NewMap returns a new hash map store
func NewMap() Store {
	return &MapStore{tab: make(map[int]uint64)}
}

// Memo gets a memoized fibonacci value
func (ms *MapStore) Memo(n int) (uint64, bool, error) {
	ms.mu.Lock()
	v, ok := ms.tab[n]
	ms.mu.Unlock()
	return v, ok, nil
}

// Memoize stores a memoized value
func (ms *MapStore) Memoize(n int, val uint64) error {
	ms.mu.Lock()
	ms.tab[n] = val
	ms.mu.Unlock()
	return nil
}

// FindLessEqual finds the highest n and value memoized value less
// than or equal to the target
func (ms *MapStore) FindLessEqual(target uint64) (int, uint64, error) {
	ms.mu.Lock()

	max := uint64(0)
	n := 0
	for k, v := range ms.tab {
		if v > max && v <= target {
			if v == target {
				return k, v, nil
			}
			max = v
			n = k
		}
	}
	ms.mu.Unlock()
	return n, max, nil
}

// Clear clears the map
func (ms *MapStore) Clear() error {
	ms.mu.Lock()
	ms.tab = make(map[int]uint64)
	ms.mu.Unlock()
	return nil
}
