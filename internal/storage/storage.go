package storage

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"proxydav/pkg/types"
)

type PersistentStore struct {
	db *badger.DB
}

func New(dataDir string) (*PersistentStore, error) {
	if dataDir == "" {
		dataDir = "./proxydavData"
	}

	opts := badger.DefaultOptions(dataDir)
	opts.Logger = nil // Disable BadgerDB logging

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	return &PersistentStore{
		db: db,
	}, nil
}

func (s *PersistentStore) Close() error {
	return s.db.Close()
}

func (s *PersistentStore) GetFileEntry(path string) (*types.FileEntry, error) {
	var entry *types.FileEntry

	err := s.db.View(func(txn *badger.Txn) error {
		key := []byte("entry:" + path)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			entry = &types.FileEntry{}
			return json.Unmarshal(val, entry)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file entry: %w", err)
	}

	return entry, nil
}

func (s *PersistentStore) SetFileEntry(entry *types.FileEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal file entry: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("entry:" + entry.Path)
		return txn.Set(key, data)
	})
}

func (s *PersistentStore) DeleteFileEntry(path string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("entry:" + path)
		return txn.Delete(key)
	})
}

func (s *PersistentStore) GetAllFileEntries() ([]types.FileEntry, error) {
	var entries []types.FileEntry

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		iter := txn.NewIterator(opts)
		defer iter.Close()

		prefix := []byte("entry:")
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()
			err := item.Value(func(val []byte) error {
				var entry types.FileEntry
				if err := json.Unmarshal(val, &entry); err != nil {
					return err
				}
				entries = append(entries, entry)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all file entries: %w", err)
	}

	return entries, nil
}

func (s *PersistentStore) GetFileMetadata(url string) (*types.FileMetadata, error) {
	var metadata *types.FileMetadata

	err := s.db.View(func(txn *badger.Txn) error {
		key := []byte("metadata:" + url)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			metadata = &types.FileMetadata{}
			return json.Unmarshal(val, metadata)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return metadata, nil
}

func (s *PersistentStore) SetFileMetadata(metadata *types.FileMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal file metadata: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("metadata:" + metadata.URL)
		return txn.Set(key, data)
	})
}

func (s *PersistentStore) DeleteFileMetadata(url string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("metadata:" + url)
		return txn.Delete(key)
	})
}

func (s *PersistentStore) RunGarbageCollection() error {
	return s.db.RunValueLogGC(0.5)
}

func (s *PersistentStore) CountFileEntries() (int, error) {
	count := 0

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need to count, not read values
		iter := txn.NewIterator(opts)
		defer iter.Close()

		prefix := []byte("entry:")
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			count++
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to count file entries: %w", err)
	}

	return count, nil
}

// GetConfig retrieves the configuration from the database
func (s *PersistentStore) GetConfig() (map[string]interface{}, error) {
	var config map[string]interface{}

	err := s.db.View(func(txn *badger.Txn) error {
		key := []byte("config:main")
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return config, nil
}

// SetConfig saves the configuration to the database
func (s *PersistentStore) SetConfig(config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("config:main")
		return txn.Set(key, data)
	})
}

// DeleteConfig removes the configuration from the database
func (s *PersistentStore) DeleteConfig() error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("config:main")
		return txn.Delete(key)
	})
}
