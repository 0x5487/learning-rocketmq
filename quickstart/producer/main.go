package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// Package main implements a simple producer to send message.
func main() {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"http://rmqnamesrv:9876"}),
		producer.WithRetry(2),
	)

	if err != nil {
		panic(err)
	}

	err = p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		os.Exit(1)
	}

	topic := "test"

	for i := 0; i < 10; i++ {
		msg := &primitive.Message{
			Topic:         topic,
			TransactionId: "strconv.Itoa(i)",
			Body:          []byte("Hello RocketMQ Go Client! " + strconv.Itoa(i)),
		}
		msg.WithProperty("order_id", "order_0001")

		res, err := p.SendSync(context.Background(), msg)

		if err != nil {
			fmt.Printf("send message error: %s\n", err)
		} else {
			fmt.Printf("send message success: result=%s\n", res.String())
		}
	}
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}
}
