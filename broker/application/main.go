package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/biagioPiraino/delphico/consumer/internal/consumers"
	"github.com/biagioPiraino/delphico/consumer/internal/databases"
	"github.com/biagioPiraino/delphico/consumer/internal/utils"
)

var (
	brokers []string
	config  *sarama.Config
)

func init() {
	// load env variables
	utils.LoadEnvVariables()

	// setup borker and consumers
	brokers = []string{"localhost:9094"}
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
}

func runFinanceConsumerGroup(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	consumer := consumers.NewFinanceConsumer()
	client, err := sarama.NewConsumerGroup(brokers, consumer.GroupId, config)
	if err != nil {
		log.Fatalf("Error while initialising finance consumer: %v", err)
	}
	defer client.Close()

	for {
		err := client.Consume(ctx, consumer.Topics, consumer)

		if ctx.Err() != nil {
			// just return, delegate closing to cleanup function in consumer logic
			return
		}

		if err != nil {
			log.Printf("Finance consumer error: %v", err)
			// prevent rapid error triggering
			time.Sleep(3 * time.Second)
		}
	}
}

func main() {
	// initialise database
	db := databases.InitDelphineDatabase()
	defer db.Pool.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go runFinanceConsumerGroup(&wg, ctx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
	log.Println("Received termination signals...closing")
	cancel()

	wg.Wait()
	log.Println("All consumer are shut down")
}
