package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/biagioPiraino/delphico/scraper/internal/core/utils"
)

var (
	logPath string
)

func InitLogger() {
	file, err := createLogFile()

	if err != nil {
		log.Fatalf("Error creating or opening log file: %v\n", err)
	}

	defer closeLogFile(file)

	// remove default timestamp to respect csv format
	log.SetFlags(0)
	log.SetOutput(file)

	log.Printf("%s,%d,logger initialised", utils.Now(), os.Getpid())
	file.Sync()
}

func LogRequest(requestId string, action string) {
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error while logging, cannot open file: %v", err)
		return
	}
	defer closeLogFile(file)

	log.SetFlags(0)
	log.SetOutput(file)
	log.Printf("%s,%s,%s", utils.Now(), requestId, action)
	file.Sync()
}

func createLogFile() (*os.File, error) {
	// Create a filename for today's logging
	filename := strings.Join([]string{today() + "_scraper_requests", "csv"}, ".")

	// Define the relative path to where to store the logs
	logDir := filepath.Join("logs")

	// Ensure the logs directory exists; if not, create it
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Create file
	logPath = filepath.Join(logDir, filename)
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func closeLogFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Printf("Error closing log file %s: %v", file.Name(), err)
	}
}

func today() string {
	return time.Now().UTC().Format("2006-01-02")
}
