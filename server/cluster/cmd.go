package cluster

import (
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	pb "google.golang.org/protobuf/proto"
)

func ApplySaveTaskCommand(raft *raft.Raft, task *proto.Task) raft.ApplyFuture {
	cmd, _ := pb.Marshal(&proto.Command{
		Cmd: &proto.Command_Task{
			Task: task,
		},
	})
	return raft.Apply(cmd, 0)
}

func ApplyAckCommand(raft *raft.Raft, ack *proto.Ack) raft.ApplyFuture {
	cmd, _ := pb.Marshal(&proto.Command{
		Cmd: &proto.Command_Ack{
			Ack: ack,
		},
	})
	return raft.Apply(cmd, 0)
}
