package main

import (
	"errors"
	"fmt"
	"github.com/ardanlabs/conf/v3"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/musabgultekin/quantumscraper/urlloader"
	"github.com/musabgultekin/quantumscraper/worker"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

var build = "develop"

func main() {
	if err := run(); err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Crawler struct {
			Concurrency int `conf:"default:100"`
		}
		UrlList struct {
			URL       string `conf:"default:https://tranco-list.eu/download/Z249G/full"`
			CachePath string `conf:"default:data/url_cache.csv"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "MIT",
		},
	}

	const prefix = "SCRAPER"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Initialization

	nsqdServer, err := storage.NewNSQDServer()
	if err != nil {
		return fmt.Errorf("start nsqd embedded server: %w", err)
	}
	queue, err := storage.NewQueue(path.Join("data/visited_urls"))
	if err != nil {
		return fmt.Errorf("visited url storage creation: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	// Queue URLs
	go func() {
		if err := startQueueingURLs(cfg.UrlList.URL, cfg.UrlList.CachePath, queue); err != nil {
			log.Fatal(err)
		}
	}()

	// Start workers
	consumer, err := worker.StartWorkers(cfg.Crawler.Concurrency, queue)
	if err != nil {
		return fmt.Errorf("worker process: %w", err)
	}

	// -------------------------------------------------------------------------
	// Shutdown

	// Wait until SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-shutdown:
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
			log.Println("URL Loader end of file")
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
