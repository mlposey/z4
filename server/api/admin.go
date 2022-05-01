package api

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/server/cluster"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Admin implements the gRPC Admin service.
type Admin struct {
	proto.UnimplementedAdminServer
	raft          *raft.Raft
	handle        *cluster.LeaderHandle
	serverID      string
	advertiseAddr string
}

func NewAdmin(raft *raft.Raft, cfg cluster.PeerConfig, handle *cluster.LeaderHandle) *Admin {
	return &Admin{
		raft:          raft,
		serverID:      cfg.ID,
		advertiseAddr: cfg.AdvertiseAddr,
		handle:        handle,
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

func (a *Admin) GetNamespace(ctx context.Context, req *proto.GetNamespaceRequest) (*proto.Namespace, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNamespace not implemented")
}

func (a *Admin) UpdateNamespace(ctx context.Context, req *proto.UpdateNamespaceRequest) (*proto.Namespace, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNamespace not implemented")
}

func (a *Admin) GetClusterInfo(
	ctx context.Context,
	req *proto.GetClusterInfoRequest,
) (*proto.ClusterInfo, error) {
	var members []*proto.Server
	config := a.raft.GetConfiguration()
	if err := config.Error(); err != nil {
		telemetry.Logger.Error("failed to fetch server list", zap.Error(err))
	} else {
		for _, server := range config.Configuration().Servers {
			members = append(members, &proto.Server{
				Id:      string(server.ID),
				Address: string(server.Address),
			})
		}
	}

	return &proto.ClusterInfo{
		ServerId:      a.serverID,
		LeaderAddress: a.handle.LeaderAddress(),
		Members:       members,
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
