package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var bootstrapServers = flag.String("bootstrap-servers", "localhost", "Kafka server(s)")
var groupId = flag.String("group-id", "consumer01", "Kafka consumer id")
var topics = flag.String("topics", "hello-world", "Kafka topic(s) to consume messages from")
var lookupInterval = flag.Int("lookup-interval", 1000, "Rate consumer lag wil be obtained (MilliSeconds, default: 1s)")
var c *kafka.Consumer

func main() {

	flag.Parse()

	messageConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": *bootstrapServers,
		"group.id":          *groupId,
		"auto.offset.reset": "earliest",
	})
	exitOnError(err)
	c = messageConsumer
	defer c.Close()
	log.Println("Connected to KafKa: ", *bootstrapServers)

	ticker := time.NewTicker(time.Duration(*lookupInterval) * time.Millisecond)
	osSignalChan := make(chan os.Signal)
	signal.Notify(osSignalChan, syscall.SIGTERM)
	signal.Notify(osSignalChan, syscall.SIGSTOP)
	signal.Notify(osSignalChan, syscall.SIGKILL)

	log.Println("[SigObserver] Observing os signals...")
	for {
		select {
		case <-osSignalChan:
			return
		case <-ticker.C:
			log.Printf("Obtain Consumer Lag for: %s/%s\n", *groupId, *topics)
			if lag, err := Backlog(); err == nil {
				log.Printf("Lag: %+v\n", lag)
			} else {
				log.Println("Unable to obtain Consumer Lag, reason: ", err)
			}
		}
	}
}

func exitOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func Backlog() (Lag, error) {

	var sum int

	//Get the assigned partitions.
	topicPartitions, err := c.Assignment()
	if err != nil {
		return Lag{}, err
	}

	// Get the offset for each partition assigned to this consumer instance
	topicPartitions, err = c.Committed(topicPartitions, 5000)
	if err != nil {
		return Lag{}, err
	}

	lag := Lag{}

	//Calculates the difference per partition of the consumer lag
	var l, highOffset int64
	for i := range topicPartitions {
		l, highOffset, err = c.QueryWatermarkOffsets(*topicPartitions[i].Topic, topicPartitions[i].Partition, 5000)
		if err != nil {
			return Lag{}, err
		}
		lag.Topic = *topicPartitions[i].Topic
		offset := int64(topicPartitions[i].Offset)
		if topicPartitions[i].Offset == kafka.OffsetInvalid {
			offset = l
		}
		//save information from the partition
		part := ConsumerInfo{
			Lag:       int(highOffset - offset),
			Partition: topicPartitions[i].Partition,
			Topic:     *topicPartitions[i].Topic,
			Offset:    int64(topicPartitions[i].Offset),
		}
		sum = sum + int(highOffset-offset)
		lag.Info = append(lag.Info, part)
	}
	lag.Lag = sum
	return lag, nil
}
