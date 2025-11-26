package app

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
)

type Config struct {
	BrokerAddress  string
	ConsumerConfig *sarama.Config
}

type App struct {
	config Config
}

func NewApp(config Config) *App {
	return &App{
		config: config,
	}
}

func (a *App) Run() {
	// initialise database
	db, err := databases.InitialisePool()
	if err != nil {
		log.Fatalf("%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// happy path
	defer func() {
		cancel() // cancelling context and close database on exit
		if err := db.Pool.Close(); err != nil {
			log.Fatalf("%v", err)
		} else {
			fmt.Println("database closed correctly")
		}
	}()

	// signals sentinel
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigchan
		fmt.Println("captured signal. closing application")
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go a.runFinanceConsumerGroup(db, &wg, ctx)

	wg.Wait()
	fmt.Println("all consumers are shut down")
}

func (a *App) runFinanceConsumerGroup(db *databases.DelphineDb, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done() // closing wait group on exit

	consumer := consumers.NewFinanceConsumer(db)
	client, err := sarama.NewConsumerGroup([]string{a.config.BrokerAddress}, consumer.GroupId, a.config.ConsumerConfig)
	if err != nil {
		fmt.Println("error initialising consumer")
		return
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Println(fmt.Sprintf("error closing sarama consumer: %v", err))
		} else {
			fmt.Println("consumer closed correctly")
		}
	}()

	for {
		err := client.Consume(ctx, consumer.Topics, consumer)

		if ctx.Err() != nil {
			return
		}

		if err != nil {
			log.Printf("Finance consumer error: %v", err)
			// prevent rapid error triggering
			time.Sleep(3 * time.Second)
		}
	}
}
