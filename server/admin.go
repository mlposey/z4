package server

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
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

	return &proto.ClusterInfo{
		ServerId:      a.serverID,
		LeaderAddress: string(a.raft.Leader()),
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
