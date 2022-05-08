package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	host := flag.String("t", "", "host:port of the peer to send requests to")
	peerAddress := flag.String("p", "", "host:port of the peer")
	peerID := flag.String("id", "", "id of the peer")
	namespace := flag.String("ns", "", "namespace")
	flag.Parse()

	if flag.NArg() != 1 {
		printHelp()
		return
	}

	admin, collection, err := makeClients(*host)
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
		consume(collection, *namespace)

	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("help")
}

func makeClients(host string) (proto.AdminClient, proto.QueueClient, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(host, opts...)
	if err != nil {
		return nil, nil, err
	}
	return proto.NewAdminClient(conn), proto.NewQueueClient(conn), nil
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

func consume(client proto.QueueClient, namespace string) {
	stream, err := client.Pull(context.Background())
	if err != nil {
		panic(err)
	}

	err = stream.Send(&proto.PullRequest{
		Request: &proto.PullRequest_StartReq{
			StartReq: &proto.StartStreamRequest{
				RequestId: ksuid.New().String(),
				Namespace: namespace,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	for {
		task, err := stream.Recv()
		if err != nil {
			fmt.Println(err)
			return
		}

		out, err := json.MarshalIndent(task, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))

		err = stream.Send(&proto.PullRequest{
			Request: &proto.PullRequest_Ack{
				Ack: &proto.Ack{
					Reference: &proto.TaskReference{
						Namespace: task.GetNamespace(),
						Id: &proto.TaskReference_TaskId{
							TaskId: task.GetId(),
						},
					},
				},
			},
		})
		if err != nil {
			panic(err)
		}
	}
}
