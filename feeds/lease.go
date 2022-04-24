package feeds

import (
	"fmt"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"sync"
	"time"
)

// leaseHolder manages access to a feed.
type leaseHolder struct {
	F        *Feed
	leases   map[*Lease]interface{}
	mu       sync.Mutex
	empty    interface{}
	onRemove func()
}

func newLeaseHolder(namespace string, db *storage.BadgerClient, onRemove func()) (*leaseHolder, error) {
	// TODO: Make ack deadline configurable.
	feed, err := New(namespace, db, time.Minute*5)
	if err != nil {
		return nil, fmt.Errorf("lease creation failed: %w", err)
	}

	return &leaseHolder{
		F:        feed,
		leases:   make(map[*Lease]interface{}),
		empty:    struct{}{},
		onRemove: onRemove,
	}, nil
}

// Get grants a client access to the feed.
func (lh *leaseHolder) Get() *Lease {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	lease := &Lease{
		feed:    lh,
		release: lh.remove,
	}
	lh.leases[lease] = true

	return lease
}

func (lh *leaseHolder) remove(l *Lease) {
	telemetry.Logger.Info("releasing lease",
		zap.String("namespace", lh.F.Namespace()))
	lh.mu.Lock()
	delete(lh.leases, l)
	lh.mu.Unlock()

	lh.onRemove()
}

// Close frees all leases and closes the feed.
//
// This method should be called when there are no active leases.
func (lh *leaseHolder) Close() error {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	for lease, _ := range lh.leases {
		delete(lh.leases, lease)
	}
	return lh.F.Close()
}

// ActiveCount is the number of open leases on the feed.
func (lh *leaseHolder) ActiveCount() int {
	return len(lh.leases)
}

// Lease is a handle on a task feed.
//
// Release should be called as soon as the lease is no longer needed.
type Lease struct {
	feed    *leaseHolder
	release func(lease *Lease)
}

func (l *Lease) Release() {
	l.feed.remove(l)
}

// Feed returns the feed being leased.
//
// The feed object must not be used after Release is called.
func (l *Lease) Feed() *Feed {
	return l.feed.F
}
