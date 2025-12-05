package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	in_kafka "github.com/biagioPiraino/delphico/scraper/internal/adapters/producers/kafka"
	"github.com/biagioPiraino/delphico/scraper/internal/core/domain"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/testcontainers/testcontainers-go"
	tc_kafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

// integration tests to assess confluent client communicate with kafka queue
func createTopicWithAdminClient(ctx context.Context, address string, topic string) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": address,
	})
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer admin.Close()

	topics := []kafka.TopicSpecification{{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}}

	results, err := admin.CreateTopics(ctx, topics, kafka.SetAdminOperationTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed creating topic: %w", err)
	}

	for _, res := range results {
		if res.Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("topic creation failed for %s: %s", res.Topic, res.Error.String())
		}
	}

	return nil
}

func setupKafkaContainer(ctx context.Context, topic string) (*tc_kafka.KafkaContainer, error) {
	kafkaContainer, err := tc_kafka.Run(ctx,
		"confluentinc/cp-kafka:7.4.0",
		tc_kafka.WithClusterID("test-cluster"),
		testcontainers.WithEnv(map[string]string{
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE": "false",
		}),
	)

	if err != nil {
		return nil, err
	}

	brokerURL, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		return nil, err
	}

	brokerAddress := brokerURL[0]

	if err := createTopicWithAdminClient(ctx, brokerAddress, topic); err != nil {
		return kafkaContainer, err
	}

	return kafkaContainer, nil
}

func createMockArticle() domain.Article {
	return domain.Article{
		Domain:    domain.FinanceDomain.String(),
		Url:       "https://foo.bar",
		Author:    "Foo Bar",
		Published: "2025-12-05",
		Provider:  "Acme Inc.",
		Content:   "This is a finance article",
	}
}

func Test_SendMessageToKafka(t *testing.T) {
	ctx := context.Background()
	kafkaContainer, err := setupKafkaContainer(ctx, "finance")
	if err != nil {
		if kafkaContainer != nil { // at this stage the container can be initialised and error only on topic creation
			if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
				t.Errorf("failed to terminate container: %v", err) // just logging error, fatal comes afterwards
			}
		}
		t.Fatalf("failed to start container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	})

	brokerUrl, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to get broker address %v", err)
	}

	const producerId = "kafka-producer"
	producer, err := in_kafka.NewConfluentProducer([]string{brokerUrl[0]}, producerId)
	if err != nil {
		t.Fatalf("failed to initialise producer %s: %v", producerId, err)
	}
	acksChannel := make(chan kafka.Event)
	producer.SetAcksChannelForTesting(acksChannel)
	t.Cleanup(producer.Shutdown)

	producer.Run()
	article := createMockArticle()
	if err := producer.SendMessageToQueue(article); err != nil {
		t.Fatalf("failed to send message to queue %v", err)
	}

	// waiting for acks from queue to assess results
	select {
	case e := <-acksChannel:
		km := e.(*kafka.Message)
		if km.TopicPartition.Error != nil {
			t.Errorf("expected successfull delivery, got error: %v", km.TopicPartition.Error)
		} else {
			t.Log("delivery successful")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out while waiting for acknowledgments")
	}
}

func Test_SendMessageToNonExistentTopic(t *testing.T) {
	ctx := context.Background()

	// the topic is attached to an article, I expect the process to fail if the target topic do not exist in Kafka
	kafkaContainer, err := setupKafkaContainer(ctx, "another_topic")
	if err != nil {
		if kafkaContainer != nil { // at this stage the container can be initialised and error only on topic creation
			if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
				t.Errorf("failed to terminate container: %v", err) // just logging error, fatal comes afterwards
			}
		}
		t.Fatalf("failed to start container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	})

	brokerUrl, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to get broker address %v", err)
	}

	const producerId = "kafka-producer"
	producer, err := in_kafka.NewConfluentProducer([]string{brokerUrl[0]}, producerId)
	if err != nil {
		t.Fatalf("failed to initialise producer %s: %v", producerId, err)
	}
	acksChannel := make(chan kafka.Event)
	producer.SetAcksChannelForTesting(acksChannel)
	t.Cleanup(producer.Shutdown)

	producer.Run()
	article := createMockArticle()
	if err := producer.SendMessageToQueue(article); err != nil {
		t.Fatalf("failed to send message to queue %v", err)
	}

	// waiting for acks from queue to assess results
	select {
	case e := <-acksChannel:
		km := e.(*kafka.Message)
		if km.TopicPartition.Error != nil {
			t.Log("error expected, msg should not be sent to non existent topics")
		} else {
			t.Errorf("expected an error here, message is not supposed to be deliverd")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out while waiting for acknowledgments")
	}
}
