package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	"z4/proto"
	"z4/storage"
)

type collection struct {
	proto.UnimplementedCollectionServer
	store *storage.SimpleStore
}

func newCollection() proto.CollectionServer {
	return &collection{store: new(storage.SimpleStore)}
}

func (c *collection) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	t, err := c.store.Save(ctx, storage.TaskDefinition{
		RunTime:  req.GetDeliverAt().AsTime(),
		Metadata: req.GetMetadata(),
		Payload:  req.GetPayload(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save task: %v", err)
	}
	return &proto.Task{
		Metadata:  t.Metadata,
		Payload:   t.Payload,
		DeliverAt: timestamppb.New(t.RunTime),
		Id:        t.ID,
	}, nil
}

func (c *collection) StreamTasks(req *proto.StreamTasksRequest, stream proto.Collection_StreamTasksServer) error {
	delay := time.Millisecond * 10
	var lastRun time.Time
	for stream.Context().Err() == nil {
		now := time.Now()
		tasks, err := c.store.Get(stream.Context(), storage.TaskRange{
			Min: lastRun,
			Max: now,
		})
		lastRun = now

		if err != nil {
			return status.Errorf(codes.Internal, "failed to fetch tasks: %v", err)
		}

		fmt.Println("sending", len(tasks), "tasks to client")
		for _, task := range tasks {
			err := stream.Send(&proto.Task{
				Metadata:  task.Metadata,
				Payload:   task.Payload,
				DeliverAt: timestamppb.New(task.RunTime),
				Id:        task.ID,
			})
			if err != nil {
				return status.Errorf(codes.Internal, "failed to send tasks to client: %v", err)
			}
		}

		sleep := time.Now().Sub(lastRun)
		if sleep < delay {
			time.Sleep(delay - sleep)
		}
	}
	fmt.Println("client conn closed")
	return nil
}
