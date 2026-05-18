package KVStore

import "errors"

type KVStore struct {
	data map[string][]byte
}

func New(walFilename string) (*KVStore, error) {
	return &KVStore{
		data: make(map[string][]byte),
	}, nil
}

func (kv KVStore) Put(key string, value []byte) error {
	kv.data[key] = value
	return nil
}

func (kv KVStore) Get(key string) ([]byte, bool, error) {
	if val, ok := kv.data[key]; ok {
		return val, true, nil
	}
	return nil, false, nil
}

func (kv KVStore) Delete(key string) error {
	if _, ok := kv.data[key]; ok {
		delete(kv.data, key)
		return nil
	}
	return errors.New("key not found: " + key)
}
