package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var bootstrapServers = flag.String("bootstrap-servers", "localhost", "Kafka server(s)")
var groupId = flag.String("group-id", "consumer01", "Kafka consumer id")
var topics = flag.String("topics", "hello-world", "Kafka topic(s) to consume messages from")
var runtime = flag.Int("runtime", 300, "Time this publisher will send messages to Kafka.(Seconds, default: 5m)")
var readInterval = flag.Int("read-interval", 1000, "Rate messages will be read from topic (MilliSeconds, default: 1s)")

func main() {

	flag.Parse()

	messageConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": *bootstrapServers,
		"group.id":          *groupId,
		"auto.offset.reset": "earliest",
	})
	exitOnError(err)
	defer messageConsumer.Close()
	log.Println("Connected to KafKa: ", *bootstrapServers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*runtime)*time.Second)
	defer cancel()

	exitOnError(messageConsumer.SubscribeTopics(strings.Split(*topics, ","), nil))
	log.Println("Subsribed to: ", *topics)

	go consumeMessages(messageConsumer, ctx)

	<-ctx.Done()
}

func consumeMessages(consumer *kafka.Consumer, ctx context.Context) {

	log.Println("Consuming messages...")
	for {
		msg, err := consumer.ReadMessage(time.Second)
		if err == nil {
			log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			log.Printf("Consumer error: %v\n", err)
		}
		log.Println("Consuming....")

		if len(ctx.Done()) > 0 {
			return
		}
		time.Sleep(time.Duration(*readInterval) * time.Millisecond)
	}
}

func exitOnError(err error) {
	if err != nil {
		panic(err)
	}
}
