package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
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
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		panic(err)
	}
}

type DemoListener struct {
	localTrans       *sync.Map
	transactionIndex int32
}

func NewDemoListener() *DemoListener {
	return &DemoListener{
		localTrans: new(sync.Map),
	}
}

func (dl *DemoListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	nextIndex := atomic.AddInt32(&dl.transactionIndex, 1)
	fmt.Printf("nextIndex: %v for transactionID: %v\n", nextIndex, msg.TransactionId)
	status := nextIndex % 3
	dl.localTrans.Store(msg.TransactionId, primitive.LocalTransactionState(status+1))

	fmt.Printf("dl")
	return primitive.UnknowState
}

func (dl *DemoListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	fmt.Printf("%v msg transactionID : %v\n", time.Now(), msg.TransactionId)
	v, existed := dl.localTrans.Load(msg.TransactionId)
	if !existed {
		fmt.Printf("unknow msg: %v, return Commit", msg)
		return primitive.CommitMessageState
	}
	state := v.(primitive.LocalTransactionState)
	switch state {
	case 1:
		fmt.Printf("checkLocalTransaction COMMIT_MESSAGE: %v\n", msg)
		return primitive.CommitMessageState
	case 2:
		fmt.Printf("checkLocalTransaction ROLLBACK_MESSAGE: %v\n", msg)
		return primitive.RollbackMessageState
	case 3:
		fmt.Printf("checkLocalTransaction unknow: %v\n", msg)
		return primitive.UnknowState
	default:
		fmt.Printf("checkLocalTransaction default COMMIT_MESSAGE: %v\n", msg)
		return primitive.CommitMessageState
	}
}

func runTx(ctx context.Context) error {

	p, _ := rocketmq.NewTransactionProducer(
		NewDemoListener(),
		producer.WithNameServer([]string{"http://namesrv:9876"}),
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"namesrv:9876"})),
		producer.WithRetry(1),
		producer.WithQueueSelector(producer.NewHashQueueSelector()),
	)
	err := p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s\n", err.Error())
		os.Exit(1)
	}

	for i := 0; i < 10; i++ {
		res, err := p.SendMessageInTransaction(context.Background(),
			primitive.NewMessage("TopicTest5", []byte("Hello RocketMQ again "+strconv.Itoa(i))))

		if err != nil {
			fmt.Printf("send message error: %s\n", err)
		} else {
			fmt.Printf("send message success: result=%s\n", res.String())
		}
	}

	time.Sleep(5 * time.Minute)
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}

	return nil
}

func run(ctx context.Context) error {

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

	// sync: 1萬筆約7.8s
	// sync: 1萬筆, 每兩筆批次, 4.576066844s
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
		msg.WithTag("order.created")

		msgs = append(msgs, msg)

		if i >= 2 && i%2 == 0 {
			err := p.SendAsync(context.Background(), func(ctx context.Context, result *primitive.SendResult, e error) {
				if e != nil {
					fmt.Printf("receive message error: %s\n", err)
				}
			}, msgs...)

			if err != nil {
				fmt.Printf("send message error: %s\n", err)
			}

			// _, err = p.SendSync(ctx, msgs...)
			// if err != nil {
			// 	fmt.Printf("send message error: %s\n", err)
			// }
		}

		msgs = msgs[:0]
	}

	fmt.Println("duration:", time.Since(start))
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}

	return nil
}
