package server

import "github.com/hashicorp/raft"

type leaderStatusChecker struct {
	raft          *raft.Raft
	leaderUpdates <-chan bool
	isLeader      bool
}

func newStatusChecker(raft *raft.Raft, selfServerID string) *leaderStatusChecker {
	checker := &leaderStatusChecker{
		raft:          raft,
		leaderUpdates: raft.LeaderCh(),
	}
	checker.detectLeader(selfServerID)
	go checker.trackLeaderChanges()
	return checker
}

func (lsc *leaderStatusChecker) detectLeader(selfServerID string) {
	lsc.isLeader = lsc.getLeaderServerID() == selfServerID
}

func (lsc *leaderStatusChecker) getLeaderServerID() string {
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

func (lsc *leaderStatusChecker) trackLeaderChanges() {
	for update := range lsc.leaderUpdates {
		lsc.isLeader = update
	}
}

func (lsc *leaderStatusChecker) IsLeader() bool {
	// TODO: Write test for this.
	return lsc.isLeader
}
