package main

import (
	"fmt"
	"github.com/musabgultekin/quantumscraper/domains"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/musabgultekin/quantumscraper/worker"
	"github.com/nsqio/go-nsq"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

const workerCount = 1000

func main() {
	if err := run(); err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize servers
	storage.StartNSQDEmbeddedServer()
	queue, err := storage.NewQueue(path.Join("queue_data"))
	if err != nil {
		return fmt.Errorf("visited url storage creation: %w", err)
	}

	// Queue domains
	go func() {
		if err := startQueueingDomains(queue); err != nil {
			log.Fatal(err)
		}
	}()

	// Start workers
	if err := startWorkers(queue); err != nil {
		return fmt.Errorf("worker process: %w", err)
	}

	// Wait until closed
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	return nil
}

func startQueueingDomains(visitedURLStorage *storage.Queue) error {
	domainLoader, err := domains.NewDomainLoader()
	if err != nil {
		return fmt.Errorf("domain loader: %w", err)
	}
	for {
		domain, err := domainLoader.NextDomain()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("next domain: %w", err)
		}

		targetURL := "https://" + domain
		if err := visitedURLStorage.AddURL(targetURL); err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}

	}
	return nil
}

// startConsumer start consumer and wait for messages
func startWorkers(visitedURLStorage *storage.Queue) error {
	consumerConfig := nsq.NewConfig()
	consumerConfig.MaxInFlight = 10
	consumer, err := nsq.NewConsumer(storage.NsqTopic, storage.NsqChannel, consumerConfig)
	if err != nil {
		return fmt.Errorf("nsq new consumer: %w", err)
	}

	consumer.AddConcurrentHandlers(worker.Worker(visitedURLStorage), 100)

	if err := consumer.ConnectToNSQD(storage.NsqServer); err != nil {
		return fmt.Errorf("connect to nsqd: %w", err)
	}

	<-consumer.StopChan
	return nil
}
