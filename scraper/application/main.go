package main

import (
	"github.com/biagioPiraino/delphico/scraper/internal/app"
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
)

func main() {
	logger.InitLogger()
	crawlerEngine := app.NewApp(app.Config{BrokerAddress: "localhost:9094"})
	crawlerEngine.Run()
}
