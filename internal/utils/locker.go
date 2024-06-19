package utils

import (
	"sync"
)

type RWLocker[T comparable] struct {
	store map[T]*sync.RWMutex // todo: sync.Map may be faster
	m     sync.Mutex
}

func NewRWLocker[T comparable]() *RWLocker[T] {
	return &RWLocker[T]{
		store: make(map[T]*sync.RWMutex),
	}
}

func (l *RWLocker[T]) RLock(key T) func() {
	lock := l.mutex(key)

	lock.RLock()

	return func() { lock.RUnlock() } // todo: monitor and cleanup unused mutexes
}

func (l *RWLocker[T]) Lock(key T) func() {
	lock := l.mutex(key)

	lock.Lock()

	return func() { lock.Unlock() }
}

// Wait will wait for all active locks to unlock, *also blocks new locks*
func (l *RWLocker[T]) Wait() {
	l.m.Lock()
	defer l.m.Unlock()

	for key, lock := range l.store {
		lock.Lock()
		delete(l.store, key)
		lock.Unlock()
	}
}

func (l *RWLocker[T]) mutex(key T) *sync.RWMutex {
	l.m.Lock()
	defer l.m.Unlock()

	lock, ok := l.store[key]
	if !ok {
		lock = &sync.RWMutex{}
		l.store[key] = lock
	}

	return lock
}
