package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mlposey/z4/pkg/z4"
	"github.com/mlposey/z4/proto"
	"go.uber.org/ratelimit"
	"os"
	"strconv"
	"strings"
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
	client, err := z4.NewClient(context.Background(), z4.ClientOptions{
		Addrs: strings.Split(os.Getenv("TARGETS"), ","),
	})
	if err != nil {
		panic(err)
	}

	const requestsToSend = 100_000_000
	done := make(chan bool)
	responses := make(chan *z4.TaskFuture, 10_000)
	count := 0

	go func() {
		for res := range responses {
			if res.Error() != nil {
				fmt.Println(res.Error())
			} else {
				count++
				if count == requestsToSend {
					done <- true
					return
				}
			}
		}
	}()
	producer, err := client.StreamingProducer(context.Background())
	if err != nil {
		panic(err)
	}

	rl := ratelimit.New(rps)
	start := time.Now()
	for i := 0; i < requestsToSend; i++ {
		rl.Take()
		responses <- producer.CreateTask(&proto.PushTaskRequest{
			RequestId: uuid.New().String(),
			Queue:     os.Getenv("QUEUE"),
			Async:     true,
			Payload:   []byte("buy eggs"),
		})
	}
	<-done
	fmt.Println("total time", time.Since(start))
}
