package main

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"

	"learning-rocketmq/quickstart/domain"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Package main implements a simple producer to send message.
func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	ctx := context.Background()

	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"http://namesrv:9876"}),
		producer.WithRetry(2),
		producer.WithQueueSelector(producer.NewHashQueueSelector()),
	)

	if err != nil {
		return err
	}

	err = p.Start()
	if err != nil {
		return err
	}

	topic := "test"

	start := time.Now()

	msgs := []*primitive.Message{}

	for i := 0; i < 10000; i++ {
		eventID := uuid.NewString()
		data := domain.CreatedOrderEvent{
			ID:            eventID,
			Sequence:      uint64(i),
			ClientOrderID: eventID,
			Market:        "BTC_USDT",
			Type:          domain.SpotOrderType_Market,
			Price:         decimal.NewFromInt(100),
			Size:          decimal.NewFromInt(100),
			Side:          domain.Side_Buy,
			TimeInForce:   domain.TIF_GTC,
			TakerFeeRate:  decimal.NewFromInt(0),
			MakerFeeRate:  decimal.NewFromInt(0),
			UserID:        123456,
			Source:        "api",
			CreatedAt:     time.Now().UTC(),
		}

		cloudEvent := cloudevents.NewEvent()
		cloudEvent.SetID(eventID)
		cloudEvent.SetSource("trade")
		cloudEvent.SetType("order.created")
		cloudEvent.SetTime(data.CreatedAt)
		err = cloudEvent.SetData(cloudevents.ApplicationJSON, data)
		if err != nil {
			return err
		}

		b, err := cloudEvent.MarshalJSON()
		if err != nil {
			return err
		}

		msg := &primitive.Message{
			Topic: topic,
			Body:  b,
		}

		msg.WithShardingKey("order_0001") // ordered message key
		msg.WithTag("CREATED_ORDER")

		msgs = append(msgs, msg)

		if i >= 10 && i%2 == 0 {
			// err := p.SendAsync(context.Background(), func(ctx context.Context, result *primitive.SendResult, e error) {
			// 	if e != nil {
			// 		fmt.Printf("receive message error: %s\n", err)
			// 	}
			// }, msgs...)

			_, err := p.SendSync(ctx, msgs...)

			if err != nil {
				fmt.Printf("send message error: %s\n", err)
			}

			msgs = msgs[:0]
		}
	}

	fmt.Println("duration:", time.Since(start))
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}

	return nil
}
