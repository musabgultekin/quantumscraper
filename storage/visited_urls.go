package storage

import (
	"fmt"
	badger "github.com/dgraph-io/badger/v3"
	"strings"
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

func (store *VisitedURLStorage) AddURL(url string) (bool, error) {
	err := store.db.Update(func(txn *badger.Txn) error {
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

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
