package main

import (
	"github.com/biagioPiraino/delphico/scraper/internal/engine"
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
)

func main() {
	logger.InitLogger()
	engine := engine.InitialiseEngine()
	engine.Run()
}
