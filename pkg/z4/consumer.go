package z4

import (
	"context"
	"errors"
	"fmt"
	"github.com/mlposey/z4/proto"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"sync"
	"sync/atomic"
)

// Consumer reads tasks as they become available for consumption.
type Consumer struct {
	stream          proto.Queue_PullClient
	ctx             context.Context
	queue           string
	acks            chan *proto.Ack
	unackedMsgCount *int64
	closed          bool
	pool            *connectionPool
	mu              sync.Mutex
}

func (c *Consumer) reconnectWithLock() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reconnect()
}

func (c *Consumer) reconnect() error {
	conn, err := c.pool.ResetConn(c.ctx)
	if err != nil {
		return err
	}
	return c.startStream(conn)
}

func (c *Consumer) startStream(conn *grpc.ClientConn) error {
	client := proto.NewQueueClient(conn)

	md := metadata.New(map[string]string{"queue": c.queue})
	ctx := metadata.NewOutgoingContext(c.ctx, md)
	stream, err := client.Pull(ctx)
	if err == nil {
		c.stream = stream
	}
	return err
}

// Consume invokes f on ready tasks.
func (c *Consumer) Consume(f func(m Message) error) error {
	err := c.startStream(c.pool.GetLeader())
	if err != nil {
		return err
	}
	go c.startAckHandler()

	for {
		select {
		case <-c.ctx.Done():
			return multierr.Combine(c.ctx.Err(), c.close())
		default:
		}

		task, err := c.stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || status.Code(err) == codes.Unavailable {
				err = c.reconnectWithLock()
				if err != nil {
					return nil
				}
				continue
			}
			return err
		}

		atomic.AddInt64(c.unackedMsgCount, 1)
		err = f(Message{
			task: task,
			acks: c.acks,
		})
		if err != nil {
			return multierr.Combine(err, c.close())
		}
	}
}

func (c *Consumer) close() error {
	c.closed = true
	return c.stream.CloseSend()
}

func (c *Consumer) startAckHandler() {
	for ack := range c.acks {
		atomic.AddInt64(c.unackedMsgCount, -1)

		if c.closed {
			if atomic.LoadInt64(c.unackedMsgCount) == 0 {
				close(c.acks)
				break
			} else {
				// Can't send ack, but can't close chan until all acks come in
				continue
			}
		}

		c.mu.Lock()
	SEND_ACK:
		err := c.stream.Send(ack)
		if err != nil {
			if status.Code(err) == codes.Unavailable {
				err = c.reconnect()
				if err == nil {
					goto SEND_ACK
				}
			}

			// TODO: Notify client that ack failed. Use a callback?
			fmt.Println("ack failed", err)
		}
		c.mu.Unlock()
	}
}

// Message represents a task that is ready for consumption.
//
// The Ack method must be invoked once the task is successfully
// processed. Failure to acknowledge a task will result in it
// being redelivered.
type Message struct {
	task *proto.Task
	acks chan<- *proto.Ack
}

func (m Message) Task() *proto.Task {
	return m.task
}

func (m Message) Ack() {
	m.acks <- &proto.Ack{
		Reference: &proto.TaskReference{
			Queue:  m.task.GetQueue(),
			TaskId: m.task.GetId(),
		},
	}
}
