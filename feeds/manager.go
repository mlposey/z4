package feeds

import (
	"github.com/mlposey/z4/storage"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
	"sync"
)

// Manager ensures that only one feed is active per requested namespace.
type Manager struct {
	feeds map[string]*managedFeed
	db    *storage.BadgerClient
	mu    sync.Mutex
}

func NewManager(db *storage.BadgerClient) *Manager {
	return &Manager{
		feeds: make(map[string]*managedFeed),
		db:    db,
	}
}

// Lease grants access to the feed for the requested namespace.
func (qm *Manager) Lease(namespace string) *Lease {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	feed, exists := qm.feeds[namespace]
	if !exists {
		feed = newManagedFeed(namespace, qm.db)
		qm.feeds[namespace] = feed
	}
	return feed.Lease()
}

func (qm *Manager) Close() error {
	// TODO: Understand implication of having leases still open.

	qm.mu.Lock()
	defer qm.mu.Unlock()
	var errs []error
	for _, feed := range qm.feeds {
		errs = append(errs, feed.F.Close())
	}
	return multierr.Combine(errs...)
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
	// TODO: Delete managedFeed from the Manager if no leases are outstanding.
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
