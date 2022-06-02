package z4

import (
	"container/list"
	"context"
	"github.com/mlposey/z4/proto"
)

// Client provides access to the features of a z4 server.
type Client struct {
	pool *connectionPool
}

func NewClient(ctx context.Context, opt ClientOptions) (*Client, error) {
	pool, err := newConnectionPool(ctx, opt.Addrs)
	if err != nil {
		return nil, err
	}
	return &Client{pool: pool}, nil
}

type ClientOptions struct {
	// Addrs is a list of host:port targets that identify members
	// of a z4 cluster.
	Addrs []string
}

func (c *Client) Close() error {
	leader := c.pool.GetLeader()
	if leader != nil {
		return leader.Close()
	}
	return nil
}

func (c *Client) UnaryProducer() *UnaryProducer {
	return &UnaryProducer{
		client: proto.NewQueueClient(c.pool.GetLeader()),
		pool:   c.pool,
	}
}

func (c *Client) StreamingProducer(ctx context.Context) (*StreamingProducer, error) {
	client := proto.NewQueueClient(c.pool.GetLeader())
	stream, err := client.PushStream(ctx)
	if err != nil {
		return nil, err
	}

	p := &StreamingProducer{
		stream:  stream,
		ctx:     ctx,
		pending: list.New(),
		pool:    c.pool,
	}
	go p.handleCallbacks()
	return p, nil
}

func (c *Client) Consumer(ctx context.Context, queue string) (*Consumer, error) {
	return &Consumer{
		queue:           queue,
		ctx:             ctx,
		acks:            make(chan *proto.Ack),
		unackedMsgCount: new(int64),
		pool:            c.pool,
	}, nil
}
