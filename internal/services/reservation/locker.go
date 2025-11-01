package reservation

import (
	"context"
	"fmt"
	"sync"
)

type Locker interface {
	Lock(ctx context.Context, key string) error
	Unlock(ctx context.Context, key string) error
}

type LocalLocker struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewLocalLocker() *LocalLocker {
	return &LocalLocker{
		locks: make(map[string]*sync.Mutex),
	}
}

func (l *LocalLocker) getOrCreateMutex(key string) *sync.Mutex {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.locks[key]; !exists {
		l.locks[key] = &sync.Mutex{}
	}

	return l.locks[key]
}

func (l *LocalLocker) Lock(ctx context.Context, key string) error {
	mu := l.getOrCreateMutex(key)
	mu.Lock()
	return nil
}

func (l *LocalLocker) Unlock(ctx context.Context, key string) error {
	l.mu.Lock()
	mutex, exists := l.locks[key]
	l.mu.Unlock()

	if !exists {
		return fmt.Errorf("no lock found for key: %s", key)
	}

	mutex.Unlock()
	return nil
}
