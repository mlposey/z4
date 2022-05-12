package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/segmentio/ksuid"
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
			fmt.Println(err)
			continue
		}
		fmt.Println(string(out))

		var ref *proto.TaskReference
		if task.GetScheduleTime() == nil {
			ref = &proto.TaskReference{
				Namespace: task.GetNamespace(),
				Id: &proto.TaskReference_Index{
					Index: task.GetIndex(),
				},
			}
		} else {
			ref = &proto.TaskReference{
				Namespace: task.GetNamespace(),
				Id: &proto.TaskReference_TaskId{
					TaskId: task.GetId(),
				},
			}
		}

		err = stream.Send(&proto.PullRequest{
			Request: &proto.PullRequest_Ack{
				Ack: &proto.Ack{
					Reference: ref,
				},
			},
		})
		if err != nil {
			fmt.Println(err)
		}
	}
}
