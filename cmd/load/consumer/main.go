package main

import (
	"fmt"
	"github.com/mlposey/z4/pkg/z4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"time"
)

func main() {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(os.Getenv("TARGET"), opts...)
	if err != nil {
		panic(err)
	}

	consumer, err := z4.NewConsumer(z4.ConsumerOptions{
		Conn:      conn,
		Namespace: os.Getenv("NAMESPACE"),
	})
	if err != nil {
		panic(err)
	}

	count := 0
	go func() {
		t := time.Tick(time.Second)
		var last int
		for range t {
			if last != count {
				fmt.Println("consumed", count, "tasks")
				last = count
			}
		}
	}()

	err = consumer.Consume(func(m z4.Message) error {
		count++
		m.Ack()
		return nil
	})
	if err != nil {
		panic(err)
	}
}
