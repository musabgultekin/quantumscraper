package main

import (
	"fmt"
	"github.com/musabgultekin/quantumscraper/domains"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/nsqio/go-nsq"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

const workerCount = 1000
const nsqServer = "localhost:4150"
const nsqTopic = "topic"
const nsqChannel = "channel"

func main() {
	if err := run(); err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize servers
	storage.StartNSQDEmbeddedServer()
	visitedURLStorage, err := storage.NewVisitedURLStorage(path.Join("visited_urls"))
	if err != nil {
		return fmt.Errorf("visited url storage creation: %w", err)
	}

	// Shared Producer
	producer, err := nsq.NewProducer(nsqServer, nsq.NewConfig())
	if err != nil {
		return fmt.Errorf("nsq new producer: %w", err)
	}

	// Queue domains
	go func() {
		if err := startQueueingDomains(visitedURLStorage, producer); err != nil {
			log.Fatal(err)
		}
	}()

	// Start workers
	if err := startWorkers(visitedURLStorage, producer); err != nil {
		return fmt.Errorf("worker process: %w", err)
	}

	// Wait until closed
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	return nil
}

func startQueueingDomains(visitedURLStorage *storage.VisitedURLStorage, producer *nsq.Producer) error {
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
		added, err := visitedURLStorage.AddURL(targetURL)
		if err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}
		if !added {
			log.Println("URL already exists:", targetURL)
			continue
		}

		if err := producer.Publish(nsqTopic, []byte(targetURL)); err != nil {
			return fmt.Errorf("producer publish: %w", err)
		}
	}
	return nil
}

// startConsumer start consumer and wait for messages
func startWorkers(visitedURLStorage *storage.VisitedURLStorage, producer *nsq.Producer) error {
	consumer, err := nsq.NewConsumer(nsqTopic, nsqChannel, nsq.NewConfig())
	if err != nil {
		return fmt.Errorf("nsq new consumer: %w", err)
	}

	consumer.AddConcurrentHandlers(worker(visitedURLStorage, producer), 1000)

	if err := consumer.ConnectToNSQD(nsqServer); err != nil {
		return fmt.Errorf("connect to nsqd: %w", err)
	}

	<-consumer.StopChan
	return nil
}

func worker(visitedURLStorage *storage.VisitedURLStorage, producer *nsq.Producer) nsq.HandlerFunc {
	return func(message *nsq.Message) error {
		log.Println(string(message.Body))

		targetURL := string(message.Body) + "&page=1"
		added, err := visitedURLStorage.AddURL(targetURL)
		if err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}
		if !added {
			log.Println("URL already exists:", targetURL)
			return nil
		}

		if err := producer.Publish(nsqTopic, []byte(targetURL)); err != nil {
			return fmt.Errorf("producer publish: %w", err)
		}

		return nil
	}
}
