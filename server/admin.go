package server

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// admin implements the gRPC Admin service.
type admin struct {
	proto.UnimplementedAdminServer
	raft     *raft.Raft
	serverID string
}

func newAdmin(raft *raft.Raft, serverID string) *admin {
	return &admin{
		raft:     raft,
		serverID: serverID,
	}
}

func (a *admin) CheckHealth(
	ctx context.Context,
	req *proto.CheckHealthRequest,
) (*proto.Status, error) {
	// TODO: Implement real health check.
	return new(proto.Status), nil
}

func (a *admin) GetClusterInfo(
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
		LeaderAddress: string(a.raft.Leader()),
		Members:       members,
	}, nil
}

func (a *admin) AddClusterMember(
	ctx context.Context,
	req *proto.AddClusterMemberRequest,
) (*emptypb.Empty, error) {
	// TODO: For now, assume this is always sent to the right node: the leader.
	// If we can detect within the server if we are the leader, we can forward
	// the request. Not sure how to do that right now

	id := raft.ServerID(req.GetMemberId())
	addr := raft.ServerAddress(req.GetMemberAddress())
	future := a.raft.AddVoter(id, addr, 0, 0)
	err := future.Error()
	return new(emptypb.Empty), status.Errorf(codes.Internal,
		"could not add member to cluster: %v", err)
}

func (a *admin) RemoveClusterMember(
	ctx context.Context,
	req *proto.RemoveClusterMemberRequest,
) (*emptypb.Empty, error) {
	// TODO: For now, assume this is always sent to the right node: the leader.
	// If we can detect within the server if we are the leader, we can forward
	// the request. Not sure how to do that right now

	id := raft.ServerID(req.GetMemberId())
	future := a.raft.RemoveServer(id, 0, 0)
	err := future.Error()
	return new(emptypb.Empty), status.Errorf(codes.Internal,
		"could not add member to cluster: %v", err)
}
