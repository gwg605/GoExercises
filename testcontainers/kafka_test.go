package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	kafka "github.com/segmentio/kafka-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

const testTopic = "test-topic"

func TestKafkaBasic(t *testing.T) {
	ctx := context.Background()
	kafkaC, err := tckafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",  // Specify the Docker image
		tckafka.WithClusterID("test-cluster"), // Optional configuration
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := kafkaC.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	brokers, err := kafkaC.Brokers(ctx)
	if err != nil {
		t.Fatalf("kafkaC.Brokers(ctx) failed. Error=%v", err)
	}

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		t.Fatalf("kafka.Dial(tcp, brokers[0]) failed. Error=%v", err)
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             testTopic,
			NumPartitions:     5,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Fatalf("conn.CreateTopics(topicConfigs...) failed. Error=%v", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        testTopic,
		RequiredAcks: -1,
		MaxAttempts:  3,
		BatchSize:    10,
		BatchTimeout: 100 * time.Millisecond,
		WriteTimeout: 1 * time.Second,
		Balancer:     &kafka.RoundRobin{},
	})
	defer func() {
		err := writer.Close()
		if err != nil {
			log.Printf("writer.Close() failed. Error=%v", err)
		}
	}()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   testTopic,
		GroupID: "Group#0",
		//Partition: 0,
		//MinBytes:  10e3,
		//MaxBytes:  10e6,
		ReadBatchTimeout: 100 * time.Millisecond,
		MaxWait:          250 * time.Millisecond,
	})
	defer func() {
		err := reader.Close()
		if err != nil {
			log.Printf("reader.Close() failed. Error=%v", err)
		}
	}()

	var msgSent atomic.Int64
	var msgReceived atomic.Int64
	wg := sync.WaitGroup{}

	for w := range 8 {
		wg.Add(1)
		go func() {
			log.Printf("[%d] Writer start", w)
			for i := range 50 {
				msg := fmt.Sprintf("Message[%d/%d]", w, i)
				err := writer.WriteMessages(ctx, kafka.Message{
					Value: []byte(msg),
				})
				if err != nil {
					log.Printf("[%d] writer.WriteMessages() failed. Error=%v", w, err)
					break
				}
				msgSent.Add(1)
				log.Printf("[%d] Write message=%s", w, msg)
				time.Sleep(50 * time.Millisecond)
			}
			log.Printf("[%d] Writer finished", w)
			wg.Done()
		}()
	}

	rctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for r := range 6 {
		wg.Add(1)
		go func() {
			var readerReceivedMsg int64
			log.Printf("[%d] Reader start", r)
			for {
				msg, err := reader.ReadMessage(rctx)
				if err != nil {
					if !errors.Is(err, context.DeadlineExceeded) {
						log.Printf("[%d] reader.ReadMessage(rctx) failed. Exit=%v", r, err)
						panic(err)
					}
					break
				}
				log.Printf("[%d] Received[%d:%d]=%s", r, msg.Partition, msg.Offset, string(msg.Value))
				err = reader.CommitMessages(rctx, msg)
				log.Printf("[%d] Commited=%s, err=%v", r, string(msg.Value), err)
				readerReceivedMsg++
			}
			msgReceived.Add(readerReceivedMsg)
			log.Printf("[%d] Reader finished (%d)", r, readerReceivedMsg)
			wg.Done()
		}()
	}

	wg.Wait()

	if msgSent != msgReceived {
		t.Fatalf("Wrong number messages: send=%d, received=%d", msgSent.Load(), msgReceived.Load())
	}

	log.Printf("Done!")
}
