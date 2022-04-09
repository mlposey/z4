package feeds

import (
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"sync"
)

// Manager ensures that only one feed is active per requested namespace.
type Manager struct {
	leases map[string]*leaseHolder
	db     *storage.BadgerClient
	mu     sync.Mutex
}

func NewManager(db *storage.BadgerClient) *Manager {
	return &Manager{
		leases: make(map[string]*leaseHolder),
		db:     db,
	}
}

// Lease grants access to the feed for the requested namespace.
func (qm *Manager) Lease(namespace string) *Lease {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	feed, exists := qm.leases[namespace]
	if !exists {
		feed = newLeaseHolder(namespace, qm.db, func() {
			qm.cleanUpLeases(namespace)
		})
		qm.leases[namespace] = feed
	}
	return feed.Lease()
}

func (qm *Manager) cleanUpLeases(namespace string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	feed, exists := qm.leases[namespace]
	if !exists {
		return
	}
	if feed.ActiveCount() > 0 {
		return
	}

	err := feed.Close()
	if err != nil {
		telemetry.Logger.Error("failed to stop feed",
			zap.Error(err),
			zap.String("namespace", namespace))
		return
	}
	delete(qm.leases, namespace)
}

// Close releases all resources for managed feeds.
func (qm *Manager) Close() error {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	var errs []error
	for ns, mf := range qm.leases {
		errs = append(errs, mf.Close())
		delete(qm.leases, ns)
	}
	return multierr.Combine(errs...)
}
