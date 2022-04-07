package feeds

import (
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"sync"
)

type leaseHolder struct {
	F        *Feed
	leases   map[*Lease]interface{}
	mu       sync.Mutex
	empty    interface{}
	onRemove func()
}

func newLeaseHolder(namespace string, db *storage.BadgerClient, onRemove func()) *leaseHolder {
	return &leaseHolder{
		F:        New(namespace, db),
		leases:   make(map[*Lease]interface{}),
		empty:    struct{}{},
		onRemove: onRemove,
	}
}

func (lh *leaseHolder) Lease() *Lease {
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

func (lh *leaseHolder) Close() error {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	for lease, _ := range lh.leases {
		delete(lh.leases, lease)
	}
	return lh.F.Close()
}

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

func (l *Lease) Feed() *Feed {
	return l.feed.F
}
