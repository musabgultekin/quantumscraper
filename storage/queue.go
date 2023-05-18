package storage

import (
	"fmt"
	"hash"
	"hash/fnv"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/nsqio/go-nsq"
)

type Queue struct {
	db          *badger.DB
	producer    *nsq.Producer
	stopped     bool
	hasher      hash.Hash64
	workerCount int

	// lastHostLock  sync.Mutex
	// lastHost      string
	// lastHostDelay time.Duration
}

func NewQueue(path string, workerCount int) (*Queue, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	// Shared Producer
	producer, err := nsq.NewProducer(NsqServer, nsq.NewConfig())
	if err != nil {
		return nil, fmt.Errorf("nsq new producer: %w", err)
	}

	return &Queue{
		db:          db,
		producer:    producer,
		hasher:      fnv.New64a(),
		workerCount: workerCount,
	}, nil
}

func (store *Queue) StopSignal() {
	store.stopped = true
}

func (store *Queue) StopProducer() {
	store.producer.Stop()
}

func (store *Queue) CloseDB() error {
	if err := store.db.Close(); err != nil {
		return fmt.Errorf("failed to close badger db: %w", err)
	}
	return nil
}

func (store *Queue) IsStopped() bool {
	return store.stopped
}

// TODO: Canonicalize links before everything
func (store *Queue) AddURL(targetURL string) error {
	// targetURLParsed, err := url.Parse(targetURL)
	// if err != nil {
	// 	return fmt.Errorf("target url parse err: %w", err)
	// }

	err := store.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(targetURL))
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get url from badger db: %w", err)
		}
		if item != nil {
			//log.Println("URL already exists:", targetURL)
			return fmt.Errorf("url %s already exists", targetURL)
		}
		err = txn.Set([]byte(targetURL), []byte{})
		return err
	})

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		} else {
			return fmt.Errorf("store db update: %w", err)
		}
	}

	// Publish if it doesn't exist
	// Select worker based on hash of the host
	// selectedTopicName := NsqTopic + strconv.Itoa(store.selectAppropriateWorkerId(targetURLParsed.Host))
	selectedTopicName := NsqTopic

	// Add delay to the host (ratelimiting)
	// NOTE: This is assuming that all URLs are ordered by their hostnames
	// store.lastHostLock.Lock()
	// {
	// 	if store.lastHost != targetURLParsed.Host {
	// 		store.lastHost = targetURLParsed.Host
	// 		store.lastHostDelay = 0
	// 	}
	// 	store.lastHostDelay += time.Second
	// 	if store.lastHostDelay > time.Hour {
	// 		store.lastHostDelay = time.Hour // FIXME: Because NSQD has 1 hour limit for deferred messages, find a way to overcome this.
	// 	}
	// }
	// store.lastHostLock.Unlock()

	// time.Sleep(time.Millisecond * 10)

	// if err := store.producer.DeferredPublish(selectedTopicName, store.lastHostDelay, []byte(targetURL)); err != nil {
	if err := store.producer.Publish(selectedTopicName, []byte(targetURL)); err != nil {
		return fmt.Errorf("producer publish: %w", err)
	}

	return nil
}

// Randomly assign a worker for the given host name.
// Use persistent hashing mechanism to find a persistently random worker id.
// It's not perfectly random, but should be good enough since it doesn't require communication or gossiping.
func (store *Queue) selectAppropriateWorkerId(targetHost string) int {
	store.hasher.Reset()
	store.hasher.Write([]byte(targetHost))
	selectedWorker := store.hasher.Sum64() % uint64(store.workerCount)
	return int(selectedWorker)
}
