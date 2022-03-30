package feeds

import (
	"github.com/mlposey/z4/storage"
	"github.com/segmentio/ksuid"
	"sync"
)

// FeedManager ensures that only one feed is active per requested namespace.
type FeedManager struct {
	queues map[string]*managedFeed
	db     *storage.BadgerClient
	mu     sync.Mutex
}

func NewFeedManager(db *storage.BadgerClient) *FeedManager {
	return &FeedManager{
		queues: make(map[string]*managedFeed),
		db:     db,
	}
}

// Lease grants access to the feed for the requested namespace.
func (qm *FeedManager) Lease(namespace string) *Lease {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	feed, exists := qm.queues[namespace]
	if !exists {
		feed = newManagedFeed(namespace, qm.db)
		qm.queues[namespace] = feed
	}
	return feed.Lease()
}

type managedFeed struct {
	F      *Feed
	leases map[string]*Lease
	mu     sync.Mutex
}

func newManagedFeed(namespace string, db *storage.BadgerClient) *managedFeed {
	return &managedFeed{
		F:      New(namespace, db),
		leases: make(map[string]*Lease),
	}
}

func (mq *managedFeed) Lease() *Lease {
	lease := &Lease{
		ID:      ksuid.New().String(),
		feed:    mq,
		release: mq.release,
	}

	mq.mu.Lock()
	mq.leases[lease.ID] = lease
	mq.mu.Unlock()
	return lease
}

func (mq *managedFeed) release(l *Lease) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	delete(mq.leases, l.ID)
	// TODO: Delete managedFeed from the FeedManager if no leases are outstanding.
}

// Lease is a handle on a task feed.
//
// Release should be called as soon as the lease is no longer needed.
type Lease struct {
	ID      string
	feed    *managedFeed
	release func(lease *Lease)
}

func (l *Lease) Release() {
	l.feed.release(l)
}

func (l *Lease) Feed() *Feed {
	return l.feed.F
}
