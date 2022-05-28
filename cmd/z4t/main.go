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
	host := flag.String("t", "", "host:port of the peer to send requests to")
	peerAddress := flag.String("p", "", "host:port of the peer")
	peerID := flag.String("id", "", "id of the peer")
	queue := flag.String("q", "", "queue")
	flag.Parse()

	if flag.NArg() != 1 {
		printHelp()
		return
	}

	admin, conn, err := makeClient(*host)
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
		consume(conn, *queue)

	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("help")
}

func makeClient(host string) (proto.AdminClient, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(host, opts...)
	if err != nil {
		return nil, nil, err
	}
	return proto.NewAdminClient(conn), conn, nil
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

func consume(conn *grpc.ClientConn, queue string) {
	consumer, err := z4.NewConsumer(z4.ConsumerOptions{
		Conn:  conn,
		Queue: queue,
	})
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
