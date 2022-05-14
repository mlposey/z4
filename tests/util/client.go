package util

import (
	"context"
	"fmt"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
)

// TODO: We should switch to using a public client once one is created.
// We need to create a Go client that users can import and use for
// interacting with z4. Once that has been created, we can directly
// use it in the functional tests rather than this custom client.

// Client is a client for the z4 server.
type Client struct {
	conn  *grpc.ClientConn
	queue proto.QueueClient
	acks  chan *proto.Ack
}

func NewClient(host string, port int) (*Client, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	target := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:  conn,
		queue: proto.NewQueueClient(conn),
	}, nil
}

// Push sends a task to the queue.
func (c *Client) Push(req *proto.PushTaskRequest) (*proto.PushTaskResponse, error) {
	return c.queue.Push(context.Background(), req)
}

// PullTasks returns a stream for consuming tasks.
func (c *Client) PullTasks(requestID, namespace string) (*PullTasksStream, error) {
	return newPullTasksStream(c.queue, requestID, namespace)
}

// Close closes the connection to the server.
func (c *Client) Close() error {
	return c.conn.Close()
}

// PullTasksStream is a handle on a task stream.
type PullTasksStream struct {
	stream    proto.Queue_PullClient
	results   chan StreamTaskResult
	acks      chan *proto.Ack
	namespace string
	requestID string
	closed    bool
}

func newPullTasksStream(queue proto.QueueClient, requestID, namespace string) (*PullTasksStream, error) {
	stream, err := queue.Pull(context.Background())
	if err != nil {
		return nil, err
	}

	return &PullTasksStream{
		stream:    stream,
		acks:      make(chan *proto.Ack),
		namespace: namespace,
		requestID: requestID,
	}, nil
}

func (s *PullTasksStream) Close() error {
	s.closed = true
	return nil
}

// Listen returns a channel that receives tasks from the server.
func (s *PullTasksStream) Listen() (<-chan StreamTaskResult, error) {
	if s.results != nil {
		return s.results, nil
	}
	s.results = make(chan StreamTaskResult)

	err := s.startStream()
	if err != nil {
		return nil, err
	}
	go s.startAckHandler()
	return s.results, nil
}

func (s *PullTasksStream) startStream() error {
	err := s.stream.Send(&proto.PullRequest{
		Request: &proto.PullRequest_StartReq{
			StartReq: &proto.StartStreamRequest{
				RequestId: s.requestID,
				Namespace: s.namespace,
			},
		},
	})
	if err != nil {
		return err
	}

	go func() {
		for !s.closed {
			task, err := s.stream.Recv()
			s.results <- StreamTaskResult{
				Error: err,
				Task:  task,
				acks:  s.acks,
			}

			if err == io.EOF {
				break
			}
		}
	}()
	return nil
}

func (s *PullTasksStream) startAckHandler() {
	for ack := range s.acks {
		err := s.stream.Send(&proto.PullRequest{
			Request: &proto.PullRequest_Ack{
				Ack: ack,
			},
		})

		if err != nil {
			fmt.Println("ack failed", err)
		}
	}
}

// StreamTaskResult is a task sent from the server.
type StreamTaskResult struct {
	Error error
	Task  *proto.Task
	acks  chan<- *proto.Ack
}

// Ack acknowledges the task so that the server does not send it again.
func (str StreamTaskResult) Ack() {
	str.acks <- &proto.Ack{
		Reference: &proto.TaskReference{
			Namespace: str.Task.GetNamespace(),
			TaskId:    str.Task.GetId(),
		},
	}
}
