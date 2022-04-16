package cluster

import (
	"github.com/hashicorp/raft"
)

type leaderObserver func(addr string)

type LeaderTracker struct {
	raft          *raft.Raft
	leaderUpdates <-chan bool
	isLeader      bool
	observers     []leaderObserver
}

func NewTracker(raft *raft.Raft, selfServerID string) *LeaderTracker {
	tracker := &LeaderTracker{
		raft:          raft,
		leaderUpdates: raft.LeaderCh(),
	}
	tracker.detectLeader(selfServerID)
	go tracker.trackLeaderChanges()
	return tracker
}

func (lsc *LeaderTracker) detectLeader(selfServerID string) {
	lsc.isLeader = lsc.getLeaderServerID() == selfServerID
}

func (lsc *LeaderTracker) getLeaderServerID() string {
	config := lsc.raft.GetConfiguration()
	if err := config.Error(); err != nil {
		return ""
	}

	leader := lsc.raft.Leader()
	peers := config.Configuration().Servers
	for _, peer := range peers {
		if peer.Address == leader {
			return string(peer.ID)
		}
	}
	return ""
}

func (lsc *LeaderTracker) trackLeaderChanges() {
	// Track when we become leader
	go func() {
		for isLeader := range lsc.leaderUpdates {
			lsc.isLeader = isLeader
		}
	}()

	// Track when anyone becomes the leader

	observations := make(chan raft.Observation)
	observer := raft.NewObserver(observations, true, func(o *raft.Observation) bool {
		switch o.Data.(type) {
		case raft.LeaderObservation:
			return true
		default:
			return false
		}
	})
	lsc.raft.RegisterObserver(observer)

	for o := range observations {
		lo := o.Data.(raft.LeaderObservation)
		leaderAddr := string(lo.Leader)
		for _, h := range lsc.observers {
			h(leaderAddr)
		}
	}
}

// AddObserver adds a handler to be called when leadership changes.
func (lsc *LeaderTracker) AddObserver(handler leaderObserver) {
	lsc.observers = append(lsc.observers, handler)
}

func (lsc *LeaderTracker) IsLeader() bool {
	// TODO: Write test for this.
	return lsc.isLeader
}

func (lsc *LeaderTracker) LeaderAddress() string {
	return string(lsc.raft.Leader())
}
