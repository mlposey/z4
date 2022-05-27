package z4

import (
	"context"
	"errors"
	"fmt"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"io"
)

type ProducerOptions struct {
	Conn *grpc.ClientConn
}

type UnaryProducer struct {
	client proto.QueueClient
}

func NewUnaryProducer(opt UnaryProducerOptions) (*UnaryProducer, error) {
	return &UnaryProducer{
		client: proto.NewQueueClient(opt.Conn),
	}, nil
}

type UnaryProducerOptions struct {
	ProducerOptions
}

func (p *UnaryProducer) CreateTask(ctx context.Context, req *proto.PushTaskRequest) (*proto.Task, error) {
	res, err := p.client.Push(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.GetTask(), nil
}

type StreamingProducer struct {
	client   proto.QueueClient
	stream   proto.Queue_PushStreamClient
	ctx      context.Context
	callback func(res StreamResponse)
	closed   bool
}

func NewStreamingProducer(opt StreamingProducerOptions) (*StreamingProducer, error) {
	client := proto.NewQueueClient(opt.Conn)

	if opt.Ctx == nil {
		opt.Ctx = context.Background()
	}

	stream, err := client.PushStream(opt.Ctx)
	if err != nil {
		return nil, err
	}

	p := &StreamingProducer{
		client:   client,
		stream:   stream,
		ctx:      opt.Ctx,
		callback: opt.Callback,
	}
	go p.handleCallbacks()
	return p, nil
}

type StreamingProducerOptions struct {
	ProducerOptions
	Ctx      context.Context
	Callback func(res StreamResponse)
}

func (p *StreamingProducer) Close() error {
	return p.stream.CloseSend()
}

func (p *StreamingProducer) handleCallbacks() {
	for !p.closed {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		res, err := p.stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// TODO: Wait to reconnect.
				return
			}
			fmt.Println(err)
			continue
		}

		if p.callback == nil {
			continue
		}

		if res.GetTask() == nil {
			p.callback(StreamResponse{
				err: TaskCreationError{
					Status:  res.GetStatus(),
					Message: res.GetMessage(),
				},
			})
		} else {
			p.callback(StreamResponse{task: res.GetTask()})
		}
	}
}

func (p *StreamingProducer) CreateTask(req *proto.PushTaskRequest) error {
	return p.stream.Send(req)
}

type StreamResponse struct {
	task *proto.Task
	err  error
}

func (r StreamResponse) Error() error {
	return r.err
}

func (r StreamResponse) Task() *proto.Task {
	return r.task
}

type TaskCreationError struct {
	Status  uint32
	Message string
}

func (e TaskCreationError) Error() string {
	return fmt.Sprintf("task creation failed: status=%d message='%s'",
		e.Status, e.Message)
}
