package feeds

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"sync"
)

// Manager ensures that only one feed is active per requested queue.
type Manager struct {
	leases map[string]*leaseHolder
	db     *storage.PebbleClient
	mu     sync.Mutex
	raft   *raft.Raft
}

func NewManager(db *storage.PebbleClient, raft *raft.Raft) *Manager {
	return &Manager{
		leases: make(map[string]*leaseHolder),
		db:     db,
		raft:   raft,
	}
}

// Tasks provides access to ready tasks from a queue.
//
// This method automatically manages a lease on the feed.
func (qm *Manager) Tasks(queue string, handle func(tasks q.TaskStream) error) error {
	lease, err := qm.Lease(queue)
	if err != nil {
		return err
	}
	defer lease.Release()

	tasks := lease.Feed().Tasks()
	return handle(tasks)
}

// Lease grants access to the feed for the requested queue.
//
// The Tasks method should be preferred in most cases.
func (qm *Manager) Lease(queue string) (*Lease, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	lease, exists := qm.leases[queue]
	if !exists {
		var err error
		lease, err = newLeaseHolder(queue, qm.db, func() {
			qm.cleanUpLeases(queue)
		}, qm.raft)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire lease for queue %s: %w", queue, err)
		}

		qm.leases[queue] = lease
	}
	return lease.Get(), nil
}

func (qm *Manager) cleanUpLeases(queue string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	feed, exists := qm.leases[queue]
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
			zap.String("queue", queue))
		return
	}
	delete(qm.leases, queue)
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
