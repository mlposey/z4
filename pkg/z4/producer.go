package z4

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/mlposey/z4/proto"
	"io"
	"sync"
)

type UnaryProducer struct {
	client proto.QueueClient
}

func (p *UnaryProducer) CreateTask(ctx context.Context, req *proto.PushTaskRequest) (*proto.Task, error) {
	res, err := p.client.Push(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.GetTask(), nil
}

type StreamingProducer struct {
	stream  proto.Queue_PushStreamClient
	ctx     context.Context
	closed  bool
	pending *list.List
	pm      sync.Mutex
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

		p.pm.Lock()
		el := p.pending.Front()
		p.pending.Remove(el)
		p.pm.Unlock()
		future := el.Value.(StreamResponse)

		if res.GetTask() == nil {
			future.errC <- TaskCreationError{
				Status:  res.GetStatus(),
				Message: res.GetMessage(),
			}
		} else {
			future.task = res.GetTask()
			future.errC <- nil
		}
	}
}

func (p *StreamingProducer) CreateTask(req *proto.PushTaskRequest) StreamResponse {
	res := StreamResponse{errC: make(chan error, 1)}
	err := p.stream.Send(req)
	if err != nil {
		res.errC <- err
	} else {
		p.pm.Lock()
		p.pending.PushBack(res)
		p.pm.Unlock()
	}
	return res
}

type StreamResponse struct {
	task *proto.Task
	errC chan error
	err  error
	done bool
}

func (r StreamResponse) Error() error {
	if r.done {
		return r.err
	}
	r.err = <-r.errC
	r.done = true
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
