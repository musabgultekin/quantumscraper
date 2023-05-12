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
	"strings"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Println("failed", err)
		os.Exit(1)
	}
}

func run() error {
	storage.StartNSQDEmbeddedServer()

	var err error
	consumer, err := startConsumer()
	if err != nil {
		return fmt.Errorf("consumer: %w", err)
	}

	visitedURLStorage, err := storage.NewVisitedURLStorage(path.Join(os.TempDir(), "visited_urls"))
	if err != nil {
		return fmt.Errorf("visited url storage creation: %w", err)
	}

	producer, err := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	if err != nil {
		return fmt.Errorf("nsq new producer: %w", err)
	}

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
		err = visitedURLStorage.AddURL(targetURL)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				fmt.Println("already exists")
				continue
			} else {
				return fmt.Errorf("visitedURL storage")
			}
		}

		if err := producer.Publish("topic", []byte(targetURL)); err != nil {
			return fmt.Errorf("producer publish: %w", err)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	consumer.Stop()

	return nil
}

// startConsumer start consumer and wait for messages
func startConsumer() (*nsq.Consumer, error) {
	consumer, err := nsq.NewConsumer("topic", "channel", nsq.NewConfig())
	if err != nil {
		return nil, fmt.Errorf("nsq new consumer: %w", err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		fmt.Println(string(message.Body))
		return nil
	}))

	if err := consumer.ConnectToNSQD("localhost:4150"); err != nil {
		return nil, fmt.Errorf("connect to nsqd: %w", err)
	}

	return consumer, nil
}
