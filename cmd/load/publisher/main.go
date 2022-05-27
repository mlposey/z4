package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mlposey/z4/pkg/z4"
	"github.com/mlposey/z4/proto"
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

	const requestsToSend = 100_000_000
	done := make(chan bool)
	count := 0

	producer, err := z4.NewStreamingProducer(z4.StreamingProducerOptions{
		ProducerOptions: z4.ProducerOptions{
			Conn: conn,
		},
		Callback: func(res z4.StreamResponse) {
			count++
			if count == requestsToSend {
				done <- true
			}
		},
	})

	if err != nil {
		panic(err)
	}

	rl := ratelimit.New(rps)
	start := time.Now()
	for i := 0; i < requestsToSend; i++ {
		rl.Take()
		err := producer.CreateTask(&proto.PushTaskRequest{
			RequestId: uuid.New().String(),
			Namespace: os.Getenv("NAMESPACE"),
			Async:     true,
			Payload:   []byte("buy eggs"),
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	<-done
	fmt.Println("total time", time.Since(start))
}
