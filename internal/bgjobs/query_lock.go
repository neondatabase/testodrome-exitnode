package bgjobs

import (
	"sync"
	"sync/atomic"
	"time"
)

// ProjectLocker allows to have in-memory per-project communication.
// For example, it can be useful to prevent deleting a project while it's being queried.
type ProjectLocker struct {
	mu sync.Mutex
	m  map[uint]*ProjectLock
}

func NewProjectLocker() *ProjectLocker {
	return &ProjectLocker{
		m: make(map[uint]*ProjectLock),
	}
}

// Get returns a lock for the project.
func (l *ProjectLocker) Get(projectID uint) *ProjectLock {
	l.mu.Lock()
	defer l.mu.Unlock()

	lock, ok := l.m[projectID]
	if !ok {
		lock = newProjectLock()
		l.m[projectID] = lock
	}
	return lock
}

// Deletes a project lock from the map. It should be deleted only after updating the database.
func (l *ProjectLocker) Delete(projectID uint) {
	go func() {
		const databaseMaxLag = 30 * time.Second

		// postpone deletion to protect against in-flight requests
		time.Sleep(databaseMaxLag)

		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.m, projectID)
	}()
}

type ProjectLock struct {
	mu sync.RWMutex

	Deleted atomic.Bool
}

func newProjectLock() *ProjectLock {
	return &ProjectLock{}
}

func (l *ProjectLock) ExclusiveLock() func() {
	l.mu.Lock()
	return l.mu.Unlock
}

func (l *ProjectLock) TryExclusiveLock() func() {
	if !l.mu.TryLock() {
		return nil
	}
	return l.mu.Unlock
}

func (l *ProjectLock) SharedLock() func() {
	l.mu.RLock()
	return l.mu.RUnlock
}

func (l *ProjectLock) TrySharedLock() func() {
	if !l.mu.TryRLock() {
		return nil
	}
	return l.mu.RUnlock
}
