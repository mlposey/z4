package cluster

import (
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
	conn     *grpc.ClientConn
	client   proto.CollectionClient
	tracker  *LeaderTracker
	grpcPort int
	mu       sync.RWMutex
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
	if leaderAddress == "" {
		return nil
	}
	lh.mu.Lock()
	lh.mu.Unlock()

	if lh.conn != nil {
		err := lh.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close leader connection: %w", err)
		}
		lh.client = nil
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
	lh.client = proto.NewCollectionClient(lh.conn)
	return nil
}

func (lh *LeaderHandle) IsLeader() bool {
	return lh.tracker.IsLeader()
}

func (lh *LeaderHandle) Client() proto.CollectionClient {
	lh.mu.RLock()
	defer lh.mu.RUnlock()
	return lh.client
}
