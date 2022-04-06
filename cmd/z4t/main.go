package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	host := flag.String("t", "", "host:port of the peer to send requests to")
	peerAddress := flag.String("p", "", "host:port of the peer")
	peerID := flag.String("id", "", "id of the peer")
	flag.Parse()

	if flag.NArg() != 1 {
		printHelp()
		return
	}

	client, err := makeClient(*host)
	if err != nil {
		panic(err)
	}

	switch flag.Arg(0) {
	case "add-peer":
		addPeer(client, *peerAddress, *peerID)

	case "remove-peer":
		removePeer(client, *peerID)

	case "info":
		info(client)

	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("help")
}

func makeClient(host string) (proto.AdminClient, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(host, opts...)
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
