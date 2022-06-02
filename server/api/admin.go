package api

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/server/cluster"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"strconv"
	"strings"
)

// Admin implements the gRPC Admin service.
type Admin struct {
	proto.UnimplementedAdminServer
	raft          *raft.Raft
	handle        *cluster.LeaderHandle
	fm            *feeds.Manager
	serverID      string
	advertiseAddr string
	ids           *storage.IDGenerator
	sqlPort       int32
	queuePort     int32
}

func NewAdmin(
	raft *raft.Raft,
	cfg cluster.PeerConfig,
	handle *cluster.LeaderHandle,
	fm *feeds.Manager,
	ids *storage.IDGenerator,
) *Admin {
	return &Admin{
		raft:          raft,
		serverID:      cfg.ID,
		advertiseAddr: cfg.AdvertiseAddr,
		handle:        handle,
		fm:            fm,
		ids:           ids,
		sqlPort:       int32(cfg.SQLPort),
		queuePort:     int32(cfg.QueuePort),
	}
}

func (a *Admin) CheckHealth(
	ctx context.Context,
	req *proto.CheckHealthRequest,
) (*proto.Status, error) {
	if a.handle.LeaderAddress() == "" {
		return nil, status.Error(codes.Internal, "peer has no leader")
	}
	return new(proto.Status), nil
}

func (a *Admin) GetQueue(ctx context.Context, req *proto.GetQueueRequest) (*proto.QueueConfig, error) {
	if !a.handle.IsLeader() {
		client, err := a.handle.AdminClient()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
		}
		return client.GetQueue(ctx, req)
	}

	feed, err := a.fm.Lease(req.GetQueue())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get lease for update: %v", err)
	}
	defer feed.Release()

	return feed.Feed().Settings.S, nil
}

func (a *Admin) UpdateQueue(ctx context.Context, req *proto.UpdateQueueRequest) (*proto.QueueConfig, error) {
	if !a.handle.IsLeader() {
		client, err := a.handle.AdminClient()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
		}
		return client.UpdateQueue(ctx, req)
	}

	feed, err := a.fm.Lease(req.GetQueue().GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get lease for update: %v", err)
	}
	defer feed.Release()

	ns := feed.Feed().Settings.S
	// Important note: Do not update the value of the queue id or the
	// last delivered task.
	if req.GetQueue().GetAckDeadlineSeconds() != 0 {
		ns.AckDeadlineSeconds = req.GetQueue().GetAckDeadlineSeconds()
	}
	return ns, nil
}

func (a *Admin) GetClusterInfo(
	ctx context.Context,
	req *proto.GetClusterInfoRequest,
) (*proto.ClusterInfo, error) {
	var members []*proto.Server
	config := a.raft.GetConfiguration()
	if err := config.Error(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get configuration: %v", err)
	}

	leaderAddr := a.handle.LeaderAddress()
	var leaderID string

	for _, server := range config.Configuration().Servers {
		addr := string(server.Address)
		if addr == leaderAddr {
			leaderID = string(server.ID)
		}

		addrParts := strings.Split(addr, ":")
		if len(addrParts) != 2 {
			return nil, status.Errorf(codes.Internal, "invalid address: %s", addr)
		}

		port, err := strconv.Atoi(addrParts[1])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid port: %s", addr)
		}

		members = append(members, &proto.Server{
			Id:       string(server.ID),
			Host:     addrParts[0],
			RaftPort: int32(port),

			// TODO: Find a way to stop assuming these.
			// It's not necessarily the case that peers are
			// using the same port numbers for these.
			SqlPort:   a.sqlPort,
			QueuePort: a.queuePort,
		})
	}

	return &proto.ClusterInfo{
		ServerId: a.serverID,
		LeaderId: leaderID,
		Members:  members,
	}, nil
}

func (a *Admin) AddClusterMember(
	ctx context.Context,
	req *proto.AddClusterMemberRequest,
) (*emptypb.Empty, error) {
	if !a.handle.IsLeader() {
		client, err := a.handle.AdminClient()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
		}
		return client.AddClusterMember(ctx, req)
	}

	if req.GetMemberId() == a.serverID {
		return nil, status.Error(codes.InvalidArgument, "cannot add leader as duplicate member of cluster")
	}

	id := raft.ServerID(req.GetMemberId())
	addr := raft.ServerAddress(req.GetMemberAddress())
	future := a.raft.AddVoter(id, addr, 0, 0)
	err := future.Error()
	if err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal,
			"could not add member to cluster: %v", err)
	}
	return new(emptypb.Empty), nil
}

func (a *Admin) RemoveClusterMember(
	ctx context.Context,
	req *proto.RemoveClusterMemberRequest,
) (*emptypb.Empty, error) {
	if !a.handle.IsLeader() {
		client, err := a.handle.AdminClient()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
		}
		return client.RemoveClusterMember(ctx, req)
	}

	id := raft.ServerID(req.GetMemberId())
	future := a.raft.RemoveServer(id, 0, 0)
	err := future.Error()
	if err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal,
			"could not remove member from cluster: %v", err)
	}
	return new(emptypb.Empty), nil
}

func (a *Admin) BootstrapCluster(ctx context.Context, e *emptypb.Empty) (*emptypb.Empty, error) {
	cfg := raft.Configuration{
		Servers: []raft.Server{
			{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(a.serverID),
				Address:  raft.ServerAddress(a.advertiseAddr),
			},
		},
	}

	f := a.raft.BootstrapCluster(cfg)
	if err := f.Error(); err != nil {
		telemetry.Logger.Error("failed to bootstrap cluster",
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to bootstrap cluster: %v", err)
	}
	return new(emptypb.Empty), nil
}
