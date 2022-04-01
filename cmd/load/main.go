package main

import (
	"context"
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"math/rand"
	"sync"
	"time"
)

func main() {
	loadTestStreaming()
}

func loadTestStreaming() {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial("localhost:6355", opts...)
	if err != nil {
		panic(err)
	}
	client := proto.NewCollectionClient(conn)
	stream, err := client.CreateTaskStreamAsync(context.Background())
	if err != nil {
		panic(err)
	}

	const (
		requestsToSend = 1_000_000
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

	start := time.Now()
	for i := 0; i < requestsToSend; i++ {
		err := stream.Send(&proto.CreateTaskRequest{
			RequestId:  ksuid.New().String(),
			Namespace:  "load_test",
			Payload:    []byte("buy eggs"),
			TtsSeconds: int64(rand.Intn(3600)), // 1 hour
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	<-done
	fmt.Println("total time", time.Since(start))
}

func loadTestUnary() {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial("localhost:6355", opts...)
	if err != nil {
		panic(err)
	}
	client := proto.NewCollectionClient(conn)

	start := time.Now()
	limit := 1_000_000
	var wg sync.WaitGroup
	wg.Add(limit)
	for i := 0; i < limit; i++ {
		go func() {
			defer wg.Done()
			_, err = client.CreateTaskAsync(context.Background(), &proto.CreateTaskRequest{
				RequestId:  ksuid.New().String(),
				Namespace:  "load_test",
				Payload:    []byte("buy eggs"),
				TtsSeconds: int64(rand.Intn(3600)), // 1 hour
			})
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	wg.Wait()
	fmt.Println("total time", time.Since(start))
}
