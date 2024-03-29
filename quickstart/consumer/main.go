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
		consumer.WithGroupName("orderGroup"),
		consumer.WithNameServer([]string{"http://namesrv:9876"}),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumerOrder(true),
	)
	defer c.Shutdown()

	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "NEW", // "TagA || TagC",
	}

	err := c.Subscribe("order", selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			msg := msgs[i]

			
			fmt.Printf("topic: %s, body: %s, shardingKey: %s, tags: %s \n", msg.Topic, string(msg.Body), msg.GetShardingKey(), msg.GetTags())
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

}
