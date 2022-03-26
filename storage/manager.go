package storage

import (
	"github.com/segmentio/ksuid"
	"sync"
)

// QueueManager ensures that only one Queue is active per requested namespace.
type QueueManager struct {
	queues map[string]*managedQueue
	db     *BadgerClient
	mu     sync.Mutex
}

func NewQueueManager(db *BadgerClient) *QueueManager {
	return &QueueManager{
		queues: make(map[string]*managedQueue),
		db:     db,
	}
}

// Lease grants access to the queue for the requested namespace.
func (qm *QueueManager) Lease(namespace string) *Lease {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	queue, exists := qm.queues[namespace]
	if !exists {
		queue = newManagedQueue(namespace, qm.db)
		qm.queues[namespace] = queue
	}
	return queue.Lease()
}

type managedQueue struct {
	Q      *Queue
	leases map[string]*Lease
	mu     sync.Mutex
}

func newManagedQueue(namespace string, db *BadgerClient) *managedQueue {
	return &managedQueue{
		Q:      New(namespace, db),
		leases: make(map[string]*Lease),
	}
}

func (mq *managedQueue) Lease() *Lease {
	lease := &Lease{
		ID:      ksuid.New().String(),
		q:       mq,
		release: mq.release,
	}

	mq.mu.Lock()
	mq.leases[lease.ID] = lease
	mq.mu.Unlock()
	return lease
}

func (mq *managedQueue) release(l *Lease) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	delete(mq.leases, l.ID)
	// TODO: Delete managedQueues from the QueueManager if no leases are outstanding.
}

// Lease is a handle on a queue.
//
// Release should be called as soon as the lease is no longer needed.
type Lease struct {
	ID      string
	q       *managedQueue
	release func(lease *Lease)
}

func (l *Lease) Release() {
	l.q.release(l)
}

func (l *Lease) Queue() *Queue {
	return l.q.Q
}
