package main

import (
	"context"
	"fmt"
	"github.com/mlposey/z4/pkg/z4"
	"os"
	"time"
)

func main() {
	client, err := z4.NewClient(z4.ClientOptions{
		Addr: os.Getenv("TARGET"),
	})
	if err != nil {
		panic(err)
	}

	consumer, err := client.Consumer(context.Background(), os.Getenv("QUEUE"))
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
