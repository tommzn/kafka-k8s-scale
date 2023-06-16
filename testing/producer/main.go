package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var bootstrapServers = flag.String("bootstrap-servers", "localhost", "Kafka server(s)")
var topic = flag.String("topic", "hello-world", "Kafka topic test messages should be published to")
var publishingInterval = flag.Int("publishing-interval", 1000, "Rate messages should be published (MilliSeconds, default: 1s)")
var runtime = flag.Int("runtime", 300, "Time this publisher will send messages to Kafka.(Seconds, default: 5m)")

func main() {

	flag.Parse()

	messagePublisher, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": *bootstrapServers})
	exitOnError(err)
	defer messagePublisher.Close()
	log.Println("Connected to KafKa: ", *bootstrapServers)

	ticker := time.NewTicker(time.Duration(*publishingInterval) * time.Millisecond)
	logTicker := time.NewTicker(30 * time.Second)
	done := make(chan bool)
	var messageCount int64
	log.Println("Publish messages to: ", *topic)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				deliveryChan := make(chan kafka.Event)
				body, _ := json.Marshal(Message{Text: "This is a test message!", TimeStamp: time.Now()})
				messagePublisher.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny},
					Value:          body,
				}, deliveryChan)
				event := <-deliveryChan
				switch ev := event.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						log.Println("Event publishing error: ", ev.TopicPartition.Error)
					}
				}
				messageCount++
			case <-logTicker.C:
				log.Printf("%d messages published\n", messageCount)
			}
		}
	}()

	runtime := time.Duration(*runtime) * time.Second
	time.Sleep(runtime)
	log.Printf("Finish message publishing after: %s\n", runtime)
	ticker.Stop()
	done <- true
	log.Printf("%d messages published\n", messageCount)
	messagePublisher.Flush(5 * 1000)
	log.Println("Done")
}

func exitOnError(err error) {
	if err != nil {
		panic(err)
	}
}
