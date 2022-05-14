package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

func main() {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(os.Getenv("TARGET"), opts...)
	if err != nil {
		panic(err)
	}
	client := proto.NewQueueClient(conn)
	consume(client, os.Getenv("NAMESPACE"))
}

func consume(client proto.QueueClient, namespace string) {
	stream, err := client.Pull(context.Background())
	if err != nil {
		panic(err)
	}

	err = stream.Send(&proto.PullRequest{
		Request: &proto.PullRequest_StartReq{
			StartReq: &proto.StartStreamRequest{
				RequestId: uuid.New().String(),
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
			fmt.Println(err)
			continue
		}
		fmt.Println(string(out))

		err = stream.Send(&proto.PullRequest{
			Request: &proto.PullRequest_Ack{
				Ack: &proto.Ack{
					Reference: &proto.TaskReference{
						Namespace: task.GetNamespace(),
						TaskId:    task.GetId(),
					},
				},
			},
		})
		if err != nil {
			fmt.Println(err)
		}
	}
}
