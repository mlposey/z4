package main

import (
	"context"
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/segmentio/ksuid"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strconv"
	"time"
)

func main() {
	rps, err := strconv.Atoi(os.Getenv("RPS"))
	if err != nil {
		panic(err)
	}
	loadTestStreaming(rps)
}

func loadTestStreaming(rps int) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(os.Getenv("TARGET"), opts...)
	if err != nil {
		panic(err)
	}
	client := proto.NewQueueClient(conn)
	stream, err := client.PushStream(context.Background())
	if err != nil {
		panic(err)
	}

	const (
		requestsToSend = 10_000_000
	)

	done := make(chan bool)
	go func() {
		count := 0
		for {
			_, err := stream.Recv()
			if err != nil {
				fmt.Println(err)
				return
			}

			count++
			if count == requestsToSend {
				done <- true
			}
		}
	}()

	rl := ratelimit.New(rps)
	start := time.Now()
	for i := 0; i < requestsToSend; i++ {
		rl.Take()
		err := stream.Send(&proto.PushTaskRequest{
			RequestId: ksuid.New().String(),
			Namespace: os.Getenv("NAMESPACE"),
			Async:     true,
			Payload:   []byte("buy eggs"),
			/*Schedule: &proto.PushTaskRequest_TtsSeconds{
				TtsSeconds: int64(rand.Intn(60)), // 1 hour
			},*/
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	<-done
	fmt.Println("total time", time.Since(start))
}
