package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ardanlabs/conf/v3"
	"github.com/musabgultekin/quantumscraper/logging"
	"github.com/musabgultekin/quantumscraper/metrics"
	"github.com/musabgultekin/quantumscraper/worker"
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
	log.SetOutput(&logging.FilteredWriter{os.Stderr})

	// Enforce proxy
	if os.Getenv("PROXY_URL") == "" {
		log.Println("Please set proxy environment variable")
		os.Exit(1)
	}

	cfg := struct {
		conf.Version
		Crawler struct {
			Concurrency int `conf:"default:100"`
		}
		UrlList struct {
			URL        string `conf:"default:https://tranco-list.eu/download/Z249G/full"`
			CachePath  string `conf:"default:data/url_cache.csv"`
			ParquetDir string `conf:"default:cc-index/"`
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

	// nsqdServer, err := storage.NewNSQDServer()
	// if err != nil {
	// 	return fmt.Errorf("start nsqd embedded server: %w", err)
	// }
	// queue, err := storage.NewQueue(path.Join("data/visited_urls"), cfg.Crawler.Concurrency)
	// if err != nil {
	// 	return fmt.Errorf("visited url storage creation: %w", err)
	// }

	go metrics.StartMetricsServer()

	// -------------------------------------------------------------------------
	// App Starting

	// Queue URLs
	// go func() {
	// 	if err := urlloader.StartQueueingURLs(cfg.UrlList.URL, cfg.UrlList.CachePath, queue); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// Start queue workers
	// consumers, err := worker.StartQueueWorkers(cfg.Crawler.Concurrency, queue)
	// if err != nil {
	// 	return fmt.Errorf("worker process: %w", err)
	// }
	var workerWg sync.WaitGroup
	worker.StartWorkers(cfg.UrlList.URL, cfg.UrlList.CachePath, cfg.UrlList.ParquetDir, &workerWg, cfg.Crawler.Concurrency)

	// -------------------------------------------------------------------------
	// Shutdown

	// Wait until SIGTERM
	// shutdown := make(chan os.Signal, 1)
	// signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	// select {
	// case <-shutdown:
	// 	// case <-nsqdServer.Error():
	// 	// 	log.Println("nsqd server error! scraping will be stopped: ", err)
	// }
	// log.Println("Stopping signal received")

	// Wait until closed
	workerWg.Wait()
	// queue.StopSignal()
	// time.Sleep(time.Millisecond * 100)
	// for _, consumer := range consumers {
	// 	consumer.Stop()
	// }
	// for _, consumer := range consumers {
	// 	<-consumer.StopChan
	// }
	// queue.StopProducer()
	// nsqdServer.Stop()
	// if err := queue.CloseDB(); err != nil {
	// 	log.Println("Queue CloseDB error:", err)
	// }

	log.Panic("Scraping successfully finished")

	return nil
}
