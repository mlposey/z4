package z4

import (
	"context"
	"errors"
	"github.com/mlposey/z4/proto"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
)

type Consumer struct {
	client    proto.QueueClient
	namespace string
	ctx       context.Context
}

func NewConsumer(opt ConsumerOptions) (*Consumer, error) {
	ctx := opt.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	return &Consumer{
		client:    proto.NewQueueClient(opt.Conn),
		namespace: opt.Namespace,
		ctx:       ctx,
	}, nil
}

type ConsumerOptions struct {
	Conn *grpc.ClientConn
	// TODO: Support multiple namespaces.
	Namespace string
	Ctx       context.Context
}

func (c *Consumer) Consume(f func(m Message) error) error {
	md := metadata.New(map[string]string{"namespace": c.namespace})
	ctx := metadata.NewOutgoingContext(c.ctx, md)
	stream, err := c.client.Pull(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-c.ctx.Done():
			return multierr.Combine(c.ctx.Err(), stream.CloseSend())
		default:
		}

		task, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		err = f(Message{
			task:   task,
			stream: stream,
		})
		if err != nil {
			return multierr.Combine(err, stream.CloseSend())
		}
	}
}

type Message struct {
	task   *proto.Task
	stream proto.Queue_PullClient
}

func (m Message) Task() *proto.Task {
	return m.task
}

func (m Message) Ack() error {
	return m.stream.Send(&proto.Ack{
		Reference: &proto.TaskReference{
			Namespace: m.task.GetNamespace(),
			TaskId:    m.task.GetId(),
		},
	})
}
