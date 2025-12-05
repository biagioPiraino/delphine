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

func Test_SendMessageToKafka(t *testing.T) {
	ctx := context.Background()

	kafkaContainer, err := tc_kafka.Run(ctx,
		"confluentinc/cp-kafka:7.4.0",
		tc_kafka.WithClusterID("test-cluster"),
	)

	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			t.Errorf("failed to terminate container: %s", err)
		}
	})

	brokerURL, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to get broker URL: %v", err)
	}

	brokerAddress := brokerURL[0]
	t.Logf("kafka broker address initialised at: %s", brokerAddress)

	const topic = "finance"

	if err := createTopicWithAdminClient(ctx, brokerAddress, topic); err != nil {
		t.Fatalf("failed to create topic %s: %v", topic, err)
	}

	const producerId = "kafka-producer"
	producer, err := in_kafka.NewConfluentProducer([]string{brokerAddress}, producerId)
	if err != nil {
		t.Fatalf("failed to initialise producer %s: %v", producerId, err)
	}
	t.Cleanup(producer.Shutdown)

	producer.Run()

	article := domain.Article{
		Domain:    domain.FinanceDomain.String(),
		Url:       "https://foo.bar",
		Author:    "Foo Bar",
		Published: "2025-12-05",
		Provider:  "Acme Inc.",
		Content:   "This is a finance article",
	}

	// if no error here, we assume that msg has been delivered without
	// checking for acks since they are processed async.
	// I wanted to avoid changing businness logic of the run method only for testing
	// purposes because this would require probably a channel to ingest the acks.
	if err := producer.SendMessageToQueue(article); err != nil {
		t.Fatalf("failed to send message to queue %v", err)
	}
}
