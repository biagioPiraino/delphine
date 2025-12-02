package main

import (
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
)

func main() {
	logger.InitLogger()
	crawlerEngine := NewApp(Config{BrokerAddress: "localhost:9094"})
	crawlerEngine.Run()
}
