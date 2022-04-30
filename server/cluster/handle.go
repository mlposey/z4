package cluster

import (
	"errors"
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
	"sync"
)

type LeaderHandle struct {
	conn             *grpc.ClientConn
	collectionClient proto.CollectionClient
	adminClient      proto.AdminClient
	tracker          *LeaderTracker
	grpcPort         int
	mu               sync.RWMutex
}

func NewHandle(tracker *LeaderTracker, grpcPort int) (*LeaderHandle, error) {
	handle := &LeaderHandle{
		tracker:  tracker,
		grpcPort: grpcPort,
	}
	leader := tracker.LeaderAddress()
	err := handle.resetConn(leader)
	if err != nil {
		return nil, err
	}

	tracker.AddObserver(func(newAddr string) {
		err := handle.resetConn(newAddr)
		if err != nil {
			telemetry.Logger.Error("failed to reset leader connection",
				zap.String("new_address", newAddr),
				zap.Error(err))
		}
	})
	return handle, nil
}

func (lh *LeaderHandle) resetConn(leaderAddress string) error {
	// TODO: Don't open connection to self if leader.

	if leaderAddress == "" {
		return nil
	}
	lh.mu.Lock()
	defer lh.mu.Unlock()

	if lh.conn != nil {
		err := lh.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close leader connection: %w", err)
		}
		lh.conn = nil
		lh.collectionClient = nil
		lh.adminClient = nil
	}

	hostPort := strings.Split(leaderAddress, ":")
	if len(hostPort) != 2 {
		return fmt.Errorf("unexpected host:port for leader: %s", leaderAddress)
	}
	leaderAddress = fmt.Sprintf("%s:%d", hostPort[0], lh.grpcPort)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(leaderAddress, opts...)
	if err != nil {
		return err
	}
	lh.conn = conn
	lh.collectionClient = proto.NewCollectionClient(lh.conn)
	lh.adminClient = proto.NewAdminClient(lh.conn)
	return nil
}

func (lh *LeaderHandle) IsLeader() bool {
	return lh.tracker.IsLeader()
}

func (lh *LeaderHandle) LeaderAddress() string {
	return lh.tracker.LeaderAddress()
}

func (lh *LeaderHandle) CollectionClient() (proto.CollectionClient, error) {
	lh.mu.RLock()
	defer lh.mu.RUnlock()

	if lh.conn == nil || lh.collectionClient == nil {
		return nil, errors.New("leader not found")
	}
	return lh.collectionClient, nil
}

func (lh *LeaderHandle) AdminClient() (proto.AdminClient, error) {
	lh.mu.RLock()
	defer lh.mu.RUnlock()

	if lh.conn == nil || lh.adminClient == nil {
		return nil, errors.New("leader not found")
	}
	return lh.adminClient, nil
}
