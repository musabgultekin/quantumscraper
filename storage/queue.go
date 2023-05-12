package storage

import (
	"time"

	badger "github.com/dgraph-io/badger/v3"
)

type Queue struct {
	db *badger.DB
}

func NewQueue(dbPath string) (*Queue, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	queue := &Queue{
		db: db,
	}

	return queue, nil
}

func (q *Queue) Enqueue(url string) error {
	err := q.db.Update(func(txn *badger.Txn) error {
		item := badger.NewEntry([]byte(time.Now().Format(time.RFC3339Nano)), []byte(url))
		err := txn.SetEntry(item)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func (q *Queue) Dequeue() (string, error) {
	var url string

	err := q.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				url = string(val)
				return nil
			})

			if err != nil {
				return err
			}

			err = txn.Delete(item.Key())
			if err != nil {
				return err
			}

			break
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return url, nil
}

func (q *Queue) Close() error {
	return q.db.Close()
}
