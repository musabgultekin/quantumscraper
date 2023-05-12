package storage

import (
	"fmt"
	badger "github.com/dgraph-io/badger/v3"
)

type VisitedURLStorage struct {
	db *badger.DB
}

func NewVisitedURLStorage(path string) (*VisitedURLStorage, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &VisitedURLStorage{
		db: db,
	}, nil
}

func (store *VisitedURLStorage) Close() error {
	if err := store.db.Close(); err != nil {
		return fmt.Errorf("failed to close badger db: %w", err)
	}
	return nil
}

func (store *VisitedURLStorage) AddURL(url string) error {
	return store.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(url))
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get url from badger db: %w", err)
		}
		if item != nil {
			return fmt.Errorf("url %s already exists", url)
		}
		err = txn.Set([]byte(url), []byte{})
		return err
	})
}

func (store *VisitedURLStorage) IsVisited(url string) (bool, error) {
	var visited bool
	err := store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(url))
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get url from badger db: %w", err)
		}
		visited = item != nil
		return nil
	})
	return visited, err
}
