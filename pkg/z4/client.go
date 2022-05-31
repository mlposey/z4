package z4

import (
	"container/list"
	"context"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client provides access to the features of a z4 server.
type Client struct {
	conn *grpc.ClientConn
}

func NewClient(opt ClientOptions) (*Client, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(opt.Addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

type ClientOptions struct {
	// Addr is the host:port of the cluster leader.
	Addr string
}

func (c *Client) UnaryProducer() *UnaryProducer {
	return &UnaryProducer{client: proto.NewQueueClient(c.conn)}
}

func (c *Client) StreamingProducer(ctx context.Context) (*StreamingProducer, error) {
	client := proto.NewQueueClient(c.conn)
	stream, err := client.PushStream(ctx)
	if err != nil {
		return nil, err
	}

	p := &StreamingProducer{
		stream:  stream,
		ctx:     ctx,
		pending: list.New(),
	}
	go p.handleCallbacks()
	return p, nil
}

func (c *Client) Consumer(ctx context.Context, queue string) (*Consumer, error) {
	return &Consumer{
		client:          proto.NewQueueClient(c.conn),
		queue:           queue,
		ctx:             ctx,
		acks:            make(chan *proto.Ack),
		unackedMsgCount: new(int64),
	}, nil
}
