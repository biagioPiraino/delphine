package logger

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/core/utils"
)

type InternalLogger struct {
	writer    *csv.Writer
	logfile   *os.File
	buffer    chan []string
	syncGroup *sync.WaitGroup
}

func NewLogger(dirTgt string, filename string) (*InternalLogger, error) {
	file, err := createDirAndFile(dirTgt, filename)
	if err != nil {
		return nil, err
	}
	writer := csv.NewWriter(file)

	// launching worker to process buffered requests
	buffer := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(1)
	go loggerWorker(&wg, writer, buffer)

	fmt.Println("logger initialised correctly")
	return &InternalLogger{
		writer:    writer,
		logfile:   file,
		buffer:    buffer,
		syncGroup: &wg,
	}, nil
}

func (i *InternalLogger) LogDebug(requestId string, operation string) {
	i.buffer <- []string{utils.Now(), requestId, operation, "OK"}
}

func (i *InternalLogger) LogError(requestId string, err error) {
	fmtError := fmt.Sprintf("%v", err)
	i.buffer <- []string{utils.Now(), requestId, fmtError, "ERR"}
}

func (i *InternalLogger) Close() {
	close(i.buffer)
	i.syncGroup.Wait()
	fmt.Println("worker terminated, flushing and closing file")
	i.writer.Flush()
	if i.logfile != nil {
		if err := i.logfile.Close(); err != nil {
			log.Printf("error while closing logging file: %v\n", err)
		} else {
			fmt.Println("logging file closed correctly")
		}
	}
}

func loggerWorker(wg *sync.WaitGroup, writer *csv.Writer, buffer <-chan []string) {
	defer wg.Done()
	var requests [][]string
	var mutex sync.RWMutex
	for request := range buffer {
		if len(requests) >= 50 {
			mutex.Lock()
			flushRequestsToDisk(writer, requests)
			requests = requests[:0]
			requests = append(requests, request)
			mutex.Unlock()
		} else {
			mutex.Lock()
			requests = append(requests, request)
			mutex.Unlock()
		}
	}

	// on return, flush out remaining of requests to avoid losing data
	if len(requests) > 0 {
		flushRequestsToDisk(writer, requests)
	}
}

func flushRequestsToDisk(writer *csv.Writer, requests [][]string) {
	sort.Slice(requests, func(i, j int) bool {
		return requests[i][0] > requests[j][0]
	})

	for _, r := range requests {
		if err := writer.Write(r); err != nil {
			log.Printf("error writing on log file %v", err)
		}
	}
	writer.Flush()
}

func createDirAndFile(dirTgt string, filename string) (*os.File, error) {
	fnWithExt := strings.Join([]string{utils.Today() + "-" + filename, "csv"}, ".")

	err := os.MkdirAll(dirTgt, 0755)
	if err != nil {
		return nil, err
	}

	logPath := filepath.Join(dirTgt, fnWithExt)
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}
