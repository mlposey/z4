package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mlposey/z4/pkg/z4"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("t", "", "host:port of the peer to send requests to")
	peerAddress := flag.String("p", "", "addr:port of the peer")
	peerID := flag.String("id", "", "id of the peer")
	queue := flag.String("q", "", "queue")
	flag.Parse()

	if flag.NArg() != 1 {
		printHelp()
		return
	}

	admin, err := newAdminClient(*addr)
	if err != nil {
		panic(err)
	}

	switch flag.Arg(0) {
	case "add-peer":
		addPeer(admin, *peerAddress, *peerID)

	case "remove-peer":
		removePeer(admin, *peerID)

	case "info":
		info(admin)

	case "consume":
		consume(*addr, *queue)

	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("help")
}

func newAdminClient(addr string) (proto.AdminClient, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return proto.NewAdminClient(conn), nil
}

func addPeer(client proto.AdminClient, addr, id string) {
	_, err := client.AddClusterMember(context.Background(), &proto.AddClusterMemberRequest{
		MemberAddress: addr,
		MemberId:      id,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("peer added")
}

func removePeer(client proto.AdminClient, id string) {
	_, err := client.RemoveClusterMember(context.Background(), &proto.RemoveClusterMemberRequest{
		MemberId: id,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("peer removed")
}

func info(client proto.AdminClient) {
	i, err := client.GetClusterInfo(context.Background(), new(proto.GetClusterInfoRequest))
	if err != nil {
		panic(err)
	}

	out, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func consume(addr, queue string) {
	client, err := z4.NewClient(z4.ClientOptions{
		Addr: addr,
	})
	if err != nil {
		panic(err)
	}

	consumer, err := client.Consumer(context.Background(), queue)
	if err != nil {
		panic(err)
	}

	err = consumer.Consume(func(m z4.Message) error {
		defer m.Ack()

		out, err := json.MarshalIndent(m.Task(), "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(out))

		return nil
	})
	if err != nil {
		panic(err)
	}
}
