package main

import (
	"context"
	"fmt"
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
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
}

func runFinanceConsumerGroup(db *databases.DelphineDb, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	consumer := consumers.NewFinanceConsumer(db)
	client, err := sarama.NewConsumerGroup(brokers, consumer.GroupId, config)
	if err != nil {
		log.Fatalf("Error while initialising finance consumer: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Println(fmt.Sprintf("error closing sarama consumer: %v", err))
		}
	}()

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
	db, err := databases.InitialisePool()
	if err != nil {
		panic(err)
	}

	// closing database on defer
	defer func() {
		if err := db.Pool.Close(); err != nil {
			log.Println(fmt.Sprintf("error closing database, %v", err))
		} else {
			log.Println("Database closed successfully")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go runFinanceConsumerGroup(db, &wg, ctx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals // graceful closure in case app terminated
	log.Println("Received termination signals...closing")
	cancel()

	wg.Wait()
	log.Println("All consumer are shut down")
}
