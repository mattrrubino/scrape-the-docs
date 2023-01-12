package crawl

import "sync"

type SafeSet[K comparable] struct {
	mu sync.Mutex
	mp map[K]bool
}

func NewSafeSet[K comparable]() *SafeSet[K] {
	return &SafeSet[K]{mp: make(map[K]bool)}
}

func (ss *SafeSet[K]) Add(key K) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	status := ss.mp[key]
	ss.mp[key] = true

	return status
}

func (ss *SafeSet[K]) Remove(key K) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	status := ss.mp[key]
	delete(ss.mp, key)

	return status
}

func (ss *SafeSet[K]) Get(key K) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	status := ss.mp[key]

	return status
}
