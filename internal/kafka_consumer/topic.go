package kafka_consumer

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func EnsureTopic(broker, topic string, partitions int, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = conn.CreateTopics(
		kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("Topic %s ensured", topic)
	return nil

}
