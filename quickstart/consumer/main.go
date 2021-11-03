package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func main() {

	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("testGroup"),
		consumer.WithNameServer([]string{"http://rmqnamesrv:9876"}),

		//consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{"http://rmqnamesrv:9876"})),
	)

	err := c.Subscribe("test", consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			msg := msgs[i]

			fmt.Printf("topic: %s, transaction_id: %s, msg_tx_id:%s, body: %s, order_id: %s \n", msg.Topic, msg.TransactionId, msg.Message.TransactionId, string(msg.Body), msg.GetProperty("order_id"))
		}

		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	// Note: start after subscribe
	err = c.Start()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	time.Sleep(time.Hour)

	err = c.Shutdown()
	if err != nil {
		fmt.Printf("shutdown Consumer error: %s", err.Error())
	}
}
