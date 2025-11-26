package main

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/biagioPiraino/delphico/consumer/internal/app"
	"github.com/biagioPiraino/delphico/consumer/internal/utils"
)

func main() {
	utils.LoadEnvVariables()

	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	appConfig := app.Config{
		BrokerAddress:  "localhost:9094",
		ConsumerConfig: config,
	}

	application := app.NewApp(appConfig)
	application.Run()
}
