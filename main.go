package main

import (
	"fmt"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/musabgultekin/quantumscraper/urlloader"
	"github.com/musabgultekin/quantumscraper/worker"
	"github.com/nsqio/go-nsq"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

const concurrency = 1000

func main() {
	if err := run(); err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize servers
	nsqdServer, err := storage.NewNSQDServer()
	if err != nil {
		return fmt.Errorf("start nsqd embedded server: %w", err)
	}
	queue, err := storage.NewQueue(path.Join("data/visited_urls"))
	if err != nil {
		return fmt.Errorf("visited url storage creation: %w", err)
	}

	// Queue URLs
	go func() {
		if err := startQueueingURLs("https://tranco-list.eu/download/Z249G/full", "data/urls.csv", queue); err != nil {
			log.Fatal(err)
		}
	}()

	// Start workers
	consumer, err := startWorkers(queue)
	if err != nil {
		return fmt.Errorf("worker process: %w", err)
	}

	// Wait until SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
	case <-nsqdServer.Error():
		log.Println("nsqd server error! scraping will be stopped: ", err)
	}
	log.Println("Stopping signal received")

	// Wait until closed
	queue.StopSignal()
	time.Sleep(time.Millisecond * 100)
	consumer.Stop()
	<-consumer.StopChan
	queue.StopProducer()
	nsqdServer.Stop()
	if err := queue.CloseDB(); err != nil {
		log.Println("Queue CloseDB error:", err)
	}

	return nil
}

func startQueueingURLs(urlListURL string, urlListPath string, queue *storage.Queue) error {
	urlLoader, err := urlloader.New(urlListURL, urlListPath)
	if err != nil {
		return fmt.Errorf("url loader: %w", err)
	}
	defer urlLoader.Close()

	for {
		targetURL, err := urlLoader.Next()
		if err != nil {
			return fmt.Errorf("next domain: %w", err)
		}
		if targetURL == "" {
			break // end of file
		}
		if queue.IsStopped() {
			break
		}
		if err := queue.AddURL(targetURL); err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}

	}
	return nil
}

// startConsumer start consumer and wait for messages
func startWorkers(queue *storage.Queue) (*nsq.Consumer, error) {
	consumerConfig := nsq.NewConfig()
	consumerConfig.MaxInFlight = 100
	consumer, err := nsq.NewConsumer(storage.NsqTopic, storage.NsqChannel, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("nsq new consumer: %w", err)
	}

	consumer.AddConcurrentHandlers(worker.Worker(queue), concurrency)

	if err := consumer.ConnectToNSQD(storage.NsqServer); err != nil {
		return nil, fmt.Errorf("connect to nsqd: %w", err)
	}
	return consumer, nil
}
