package storage

import (
	"fmt"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/nsqio/go-nsq"
	"log"
	"strings"
)

type Queue struct {
	db       *badger.DB
	producer *nsq.Producer
}

func NewQueue(path string) (*Queue, error) {
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
		db:       db,
		producer: producer,
	}, nil
}

func (store *Queue) Close() error {
	if err := store.db.Close(); err != nil {
		return fmt.Errorf("failed to close badger db: %w", err)
	}
	return nil
}

func (store *Queue) AddURL(targetURL string) error {
	err := store.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(targetURL))
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get url from badger db: %w", err)
		}
		if item != nil {
			log.Println("URL already exists:", targetURL)
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

	// Publish if it doesn't exists
	if err := store.producer.Publish(NsqTopic, []byte(targetURL)); err != nil {
		return fmt.Errorf("producer publish: %w", err)
	}

	return nil
}
