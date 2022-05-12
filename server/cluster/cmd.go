package cluster

import (
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	pb "google.golang.org/protobuf/proto"
)

// ApplySaveTaskCommand applies a command to save a task to the Raft log.
func ApplySaveTaskCommand(raft *raft.Raft, task *proto.Task) raft.ApplyFuture {
	cmd, _ := pb.Marshal(&proto.Command{
		Cmd: &proto.Command_Task{
			Task: task,
		},
	})
	return raft.Apply(cmd, 0)
}

// ApplyAckCommand applies a command to save ack a task to the Raft log.
func ApplyAckCommand(raft *raft.Raft, ack *proto.Ack) raft.ApplyFuture {
	cmd, _ := pb.Marshal(&proto.Command{
		Cmd: &proto.Command_Ack{
			Ack: ack,
		},
	})
	return raft.Apply(cmd, 0)
}

// ApplyPurgeTasksCommand applies a command to purge namespace tasks to the Raft log.
func ApplyPurgeTasksCommand(raft *raft.Raft, req *proto.PurgeTasksRequest) raft.ApplyFuture {
	cmd, _ := pb.Marshal(&proto.Command{
		Cmd: &proto.Command_Purge{
			Purge: req,
		},
	})
	return raft.Apply(cmd, 0)
}
