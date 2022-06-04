package z4

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync"
)

var (
	ErrInterruptedFuture = errors.New("connection closed before future completion")
)

// UnaryProducer supports the creation of tasks using the unary rpc.
//
// This type is ideal for cases where the caller must immediately
// know whether a request succeeded or failed. Stronger consistency
// comes at the cost of slower performance. For high performance use
// cases, consider using StreamingProducer instead.
type UnaryProducer struct {
	client proto.QueueClient
	pool   *connectionPool
	mu     sync.Mutex
}

func (p *UnaryProducer) CreateTask(ctx context.Context, req *proto.PushTaskRequest) (*proto.Task, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	res, err := p.client.Push(ctx, req)
	if err != nil {
		code := status.Code(err)
		if code == codes.Unavailable {
			conn, err := p.pool.ResetConn(ctx)
			if err != nil {
				return nil, err
			}
			p.client = proto.NewQueueClient(conn)
			return p.CreateTask(ctx, req)
		}
		return nil, err
	}

	if res.GetForwardedTo() != "" {
		// TODO: use pool.SetLeader instead
		// Need to fix port in GetForwardedTo response first.
		conn, err := p.pool.ResetConn(ctx)
		if err != nil {
			fmt.Println("failed to reset conn", err)
		} else {
			p.client = proto.NewQueueClient(conn)
		}
	}
	return res.GetTask(), nil
}

// StreamingProducer supports the creation of tasks using bidirectional streaming.
//
// StreamingProducer offers better performance than UnaryProducer
// but at the cost of weaker consistency and a more complex interface.
//
// Close() should be called when the producer object is no longer needed.
type StreamingProducer struct {
	stream proto.Queue_PushStreamClient
	ctx    context.Context

	pending *list.List
	pm      sync.Mutex

	pool *connectionPool
}

func (p *StreamingProducer) reconnectWithLock() error {
	p.pm.Lock()
	defer p.pm.Unlock()
	return p.reconnect()
}

func (p *StreamingProducer) reconnect() error {
	conn, err := p.pool.ResetConn(p.ctx)
	if err != nil {
		return err
	}

	for el := p.pending.Front(); el != nil; el = el.Next() {
		future := el.Value.(*TaskFuture)
		future.set(nil, ErrInterruptedFuture)
	}
	p.pending.Init()

	client := proto.NewQueueClient(conn)
	stream, err := client.PushStream(p.ctx)
	if err != nil {
		return err
	}
	p.stream = stream
	return nil
}

// Close releases resources allocated to the producer.
//
// This method should be called when the producer is no longer needed.
func (p *StreamingProducer) Close() error {
	return p.stream.CloseSend()
}

func (p *StreamingProducer) handleCallbacks() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		res, err := p.stream.Recv()
		if err != nil {
			// TODO: If error is leadership lost, reconnect
			// rpc error: code = Internal desc = failed to save task: leadership lost while committing log
			if errors.Is(err, io.EOF) || status.Code(err) == codes.Unavailable {
				err = p.reconnectWithLock()
				if err != nil {
					return
				}
				continue
			}

			//fmt.Println(err)
			continue
		}

		p.pm.Lock()
		el := p.pending.Front()
		p.pending.Remove(el)
		p.pm.Unlock()

		future := el.Value.(*TaskFuture)
		if res.GetTask() == nil {
			future.set(nil, TaskCreationError{
				Status:  res.GetStatus(),
				Message: res.GetMessage(),
			})
		} else {
			future.set(res.GetTask(), nil)

			if res.GetForwardedTo() != "" {
				err = p.reconnectWithLock()
				if err != nil {
					return
				}
			}
		}
	}
}

// CreateTask sends a request to create a task to the stream.
//
// Because this is a streaming producer, the server response may not
// be immediately available. The returned future will provide access
// to the task (or an error upon failure) once the response is received.
func (p *StreamingProducer) CreateTask(req *proto.PushTaskRequest) *TaskFuture {
	p.pm.Lock()
	defer p.pm.Unlock()

	future := &TaskFuture{errC: make(chan error, 1)}
SEND:
	err := p.stream.Send(req)
	if err != nil {
		if status.Code(err) == codes.Unavailable {
			err = p.reconnect()
			if err == nil {
				goto SEND
			}
		}
		future.set(nil, err)
	} else {
		p.pending.PushBack(future)
	}
	return future
}

// TaskFuture is an async response to creating a task.
//
// To use objects of this type, first invoke the Error method.
// It will block until a response is received from the server.
// If the error is nil, the Task method will provide a non-nil task.
type TaskFuture struct {
	task *proto.Task
	err  error
	errC chan error
}

func (f *TaskFuture) set(task *proto.Task, err error) {
	f.task = task
	f.errC <- err
	close(f.errC)
}

func (f *TaskFuture) Error() error {
	if f.err != nil {
		return f.err
	}
	f.err = <-f.errC
	return f.err
}

func (f *TaskFuture) Task() *proto.Task {
	return f.task
}

// TaskCreationError represents an error encountered by the server
// while creating a task.
type TaskCreationError struct {
	Status  uint32
	Message string
}

func (e TaskCreationError) Error() string {
	return fmt.Sprintf("task creation failed: status=%d message='%s'",
		e.Status, e.Message)
}
