package z4

import (
	"context"
	"errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
	"sync"
)

var (
	ErrNoLeader = errors.New("leader not found")
)

type connectionPool struct {
	conn          *grpc.ClientConn
	peerAddresses []string
	term          int
	mu            sync.Mutex
}

func newConnectionPool(ctx context.Context, addrs []string) (*connectionPool, error) {
	p := &connectionPool{peerAddresses: addrs}
	_, err := p.ResetConn(ctx)
	return p, err
}

func (p *connectionPool) GetLeader() *grpc.ClientConn {
	return p.conn
}

func (p *connectionPool) SetLeader(addr string) (*grpc.ClientConn, error) {
	term := p.term
	p.mu.Lock()
	if term != p.term {
		p.mu.Unlock()
		return p.conn, nil
	}
	defer p.mu.Unlock()

	err := p.seed(addr)
	if err == nil {
		p.term++
	}
	return p.conn, err
}

func (p *connectionPool) ResetConn(ctx context.Context) (*grpc.ClientConn, error) {
	term := p.term
	p.mu.Lock()
	if term != p.term {
		p.mu.Unlock()
		return p.conn, nil
	}
	defer p.mu.Unlock()

	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	err := backoff.Retry(func() error {
		return p.detectLeader(p.peerAddresses)
	}, b)
	if err != nil {
		return nil, err
	}

	p.term++
	return p.conn, nil
}

func (p *connectionPool) detectLeader(addrs []string) error {
	for _, addr := range addrs {
		err := p.seed(addr)
		if err == nil {
			return nil
		}
	}
	return ErrNoLeader
}

func (p *connectionPool) seed(addr string) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return err
	}

	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn = conn
	admin := proto.NewAdminClient(p.conn)

	ctx := context.Background()
	info, err := admin.GetClusterInfo(ctx, new(proto.GetClusterInfoRequest))
	if err != nil {
		return err
	}

	leaderHost := strings.Split(info.GetLeaderAddress(), ":")[0]
	port := strings.Split(addr, ":")[1]
	leaderAddr := leaderHost + ":" + port

	if leaderAddr == "" {
		_ = p.conn.Close()
		p.conn = nil
		return ErrNoLeader
	} else if leaderAddr == addr {
		p.peerAddresses = nil
		for _, member := range info.GetMembers() {
			memberHost := strings.Split(member.GetAddress(), ":")[0]
			peerAddr := memberHost + ":" + port
			p.peerAddresses = append(p.peerAddresses, peerAddr)
		}
		return nil
	} else {
		_ = p.conn.Close()
		return p.seed(leaderAddr)
	}
}
