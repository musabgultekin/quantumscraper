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
	"time"
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
	if err := storage.StartNSQDEmbeddedServer(); err != nil {
		return fmt.Errorf("start nsqd embedded server: %w", err)
	}
	queue, err := storage.NewQueue(path.Join("data/visited_urls"))
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
	consumer, err := startWorkers(queue)
	if err != nil {
		return fmt.Errorf("worker process: %w", err)
	}

	// Wait until SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Stopping signal received")

	// Wait until closed
	queue.StopSignal()
	time.Sleep(time.Millisecond * 100)
	consumer.Stop()
	<-consumer.StopChan
	queue.StopProducer()
	if err := queue.CloseDB(); err != nil {
		log.Println("Queue CloseDB error:", err)
	}

	return nil
}

func startQueueingDomains(queue *storage.Queue) error {
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

		if queue.IsStopped() {
			break
		}

		targetURL := "https://" + domain
		if err := queue.AddURL(targetURL); err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}

	}
	return nil
}

// startConsumer start consumer and wait for messages
func startWorkers(queue *storage.Queue) (*nsq.Consumer, error) {
	consumerConfig := nsq.NewConfig()
	consumerConfig.MaxInFlight = 10
	consumer, err := nsq.NewConsumer(storage.NsqTopic, storage.NsqChannel, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("nsq new consumer: %w", err)
	}

	consumer.AddConcurrentHandlers(worker.Worker(queue), 10)

	if err := consumer.ConnectToNSQD(storage.NsqServer); err != nil {
		return nil, fmt.Errorf("connect to nsqd: %w", err)
	}
	return consumer, nil
}
